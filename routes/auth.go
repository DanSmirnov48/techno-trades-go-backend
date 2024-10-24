package routes

import (
	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	userManager = managers.UserManager{}
)

func (endpoint Endpoint) Login(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.LoginSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: reqData.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil || !utils.CheckPasswordHash(reqData.Password, user.Password) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	if !user.IsEmailVerified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Create Auth Tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: SuccessResponse("Login successful"),
		Data:           schemas.TokensResponseSchema{User: &user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) Logout(c *fiber.Ctx) error {
	// Remove the access token cookie
	auth.RemoveAuthCookie(c, auth.AccessToken)
	auth.RemoveAuthCookie(c, auth.RefreshToken)

	return c.Status(200).JSON(SuccessResponse("Logout successful"))
}

func (endpoint Endpoint) Register(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.RegisterUser{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	userByEmail, _ := userManager.GetByEmail(db, reqData.Email)
	if userByEmail != nil {
		data := map[string]string{
			"email": "Email already registered!",
		}
		return c.Status(422).JSON(utils.RequestErr(utils.ERR_INVALID_ENTRY, "Invalid Entry", data))
	}

	// Create User
	newUser := userManager.Create(db, reqData, false, false)
	if newUser.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Error creating user"))
	}

	// Create Otp
	otp := models.Otp{UserId: newUser.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	response := schemas.RegisterResponseSchema{
		ResponseSchema: SuccessResponse("Registration successful"),
		Data:           schemas.EmailRequestSchema{Email: newUser.Email},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) VerifyAccount(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.VerifyAccountRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: reqData.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	if user.IsEmailVerified {
		return c.Status(200).JSON(SuccessResponse("Email already verified"))
	}

	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != reqData.Otp {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	if otp.CheckExpiration() {
		return c.Status(400).JSON(utils.RequestErr(utils.ERR_EXPIRED_OTP, "Expired Otp"))
	}

	// Update User & Delete Otp
	user.IsEmailVerified = true
	db.Save(&user)
	db.Delete(&otp)

	return c.Status(200).JSON(SuccessResponse("Account verification successful"))
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
		ResponseSchema: SuccessResponse("Validate successful"),
		Data:           schemas.TokensResponseSchema{User: user, Access: accessToken},
	}

	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) Refresh(c *fiber.Ctx) error {
	reqData := schemas.RefreshTokenRequestSchema{}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	// Validate request
	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	token := reqData.Refresh
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
		ResponseSchema: SuccessResponse("Tokens refresh successful"),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}
