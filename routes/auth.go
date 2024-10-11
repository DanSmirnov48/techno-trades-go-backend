package routes

import (
	"fmt"

	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
)

var (
	validator   = utils.Validator()
	cfg         = config.GetConfig()
	userManager = managers.UserManager{}
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
	user, _ := userManager.GetByEmail(db, userLoginSchema.Email)
	if user == nil || !user.ComparePassword(userLoginSchema.Password) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	if !user.Verified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
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

	userByEmail, _ := userManager.GetByEmail(db, user.Email)
	if userByEmail != nil {
		data := map[string]string{
			"email": "Email already registered!",
		}
		return c.Status(422).JSON(utils.RequestErr(utils.ERR_INVALID_ENTRY, "Invalid Entry", data))
	}

	// Create User
	newUser, err := userManager.Create(db, user, false, false)
	if err != nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, err.Message))
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

	user, _ := userManager.GetByEmail(db, input.Email)
	if user == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	if user.Verified {
		return c.Status(200).JSON(schemas.ResponseSchema{Message: "Email already verified"}.Init())
	}

	if user.VerificationCode != input.VerificationCode {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	if err := userManager.SetAccountVerified(db, user); err != nil {
		return c.Status(err.Code).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, err.Message))
	}

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
	refreshTokenSchema := schemas.RefreshTokenRequestSchema{}

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

	user, _ := userManager.GetByEmail(db, emailSchema.Email)
	if user == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INVALID_OWNER, "User not found"))
	}

	if !user.Verified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	token, err := user.CreateMagicLogInLinkToken()
	if err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to generate magic login token"))
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

	user, err := userManager.GetByMagicLoginToken(db, token)
	if err != nil {
		return c.Status(err.Code).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, err.Message))
	}

	if !user.Verified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	userManager.ClearMagicLogin(db, user)

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
