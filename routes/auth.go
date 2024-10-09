package routes

import (
	"fmt"
	"time"

	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	validator = utils.Validator()
	cfg       = config.GetConfig()
)

func (endpoint Endpoint) Login(c *fiber.Ctx) error {
	db := endpoint.DB
	userLoginSchema := schemas.LoginSchema{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &userLoginSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(userLoginSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	// Check if the user exists and validate password
	user, _ := controllers.GetUserByEmail(db, userLoginSchema.Email)
	if user == nil || !user.ComparePassword(userLoginSchema.Password) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	// Create Auth Tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Login successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) Logout(c *fiber.Ctx) error {
	// Remove the access token cookie
	auth.RemoveAuthCookie(c, auth.AccessToken)
	auth.RemoveAuthCookie(c, auth.RefreshToken)

	response := schemas.ResponseSchema{Message: "Logout successful"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) Register(c *fiber.Ctx) error {
	db := endpoint.DB
	user := schemas.RegisterUser{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &user); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(user); err != nil {
		return c.Status(422).JSON(err)
	}

	userByEmail, _ := controllers.GetUserByEmail(db, user.Email)
	if userByEmail != nil {
		data := map[string]string{
			"email": "Email already registered!",
		}
		return c.Status(422).JSON(utils.RequestErr(utils.ERR_INVALID_ENTRY, "Invalid Entry", data))
	}

	newUser := models.User{
		ID:               uuid.New(),
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		Password:         user.Password,
		VerificationCode: utils.GetRandomInt(6),
	}

	if err := db.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not create user")
	}

	response := schemas.RegisterResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Registration successful"}.Init(),
		Data:           schemas.EmailRequestSchema{Email: newUser.Email},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) VerifyAccount(c *fiber.Ctx) error {
	db := endpoint.DB
	input := schemas.VerifyAccountRequestSchema{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &input); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(input); err != nil {
		return c.Status(422).JSON(err)
	}

	user, _ := controllers.GetUserByEmail(db, input.Email)
	if user == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	if user.VerificationCode != input.VerificationCode {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	user.Verified = true
	user.VerificationCode = 0

	db.Save(&user)

	response := schemas.ResponseSchema{Message: "Account verification successful"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) ValidateMe(c *fiber.Ctx) error {
	db := endpoint.DB

	accessToken := c.Cookies("accessToken")

	if accessToken == "" {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, "Invalid Token"))
	}

	user, err := auth.DecodeAccessToken(accessToken, db)
	if err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	c.Locals("user", user)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Validate successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: accessToken},
	}

	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) Refresh(c *fiber.Ctx) error {
	refreshTokenSchema := schemas.RefreshTokenSchema{}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &refreshTokenSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(refreshTokenSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	token := refreshTokenSchema.Refresh
	if !auth.DecodeRefreshToken(token) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, "Refresh token is invalid or expired"))
	}

	// Create Auth Tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Tokens refresh successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) SendMagicLink(c *fiber.Ctx) error {
	db := endpoint.DB
	emailSchema := schemas.EmailRequestSchema{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &emailSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(emailSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	user, _ := controllers.GetUserByEmail(db, emailSchema.Email)
	if user == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	token, err := user.CreateMagicLogInLinkToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate magic login token",
		})
	}

	db.Save(&user)

	magicLink := fmt.Sprintf("%s/login/%s", cfg.FrontendURL, token)

	response := schemas.MagicLinkLoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "MagicLink has been sent"}.Init(),
		Data:           schemas.MagicLinkResponseSchema{Link: magicLink},
	}
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) MagicLinkLogin(c *fiber.Ctx) error {
	db := endpoint.DB

	token := c.Params("token")

	var user *models.User

	if err := db.Where("magic_log_in_token = ? AND magic_log_in_token_expires > ?", token, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, "Refresh token is invalid or expired"))
		}
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	db.Model(&user).Updates(map[string]interface{}{
		"MagicLogInToken":        nil,
		"MagicLogInTokenExpires": nil,
	})

	// Create Auth Tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Login successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}
