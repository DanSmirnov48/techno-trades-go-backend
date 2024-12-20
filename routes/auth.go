package routes

import (
	"strconv"

	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/senders"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	user.Access = &access
	user.Refresh = &refresh
	db.Save(&user)

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
	db := endpoint.DB
	user := RequestUser(c)
	user.Access = nil
	user.Refresh = nil
	db.Save(user)

	// Remove the access token cookie
	auth.RemoveAuthCookie(c, auth.AccessToken)
	auth.RemoveAuthCookie(c, auth.RefreshToken)

	return c.Status(200).JSON(SuccessResponse("Logout successful"))
}

func (endpoint Endpoint) Register(c *fiber.Ctx) error {
	db := endpoint.DB
	data := schemas.RegisterUser{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := utils.ConvertStructData(data, models.User{}).(*models.User)
	// Validate email uniqueness
	db.Take(&user, models.User{Email: user.Email})
	if user.ID != uuid.Nil {
		data := map[string]string{
			"email": "Email already taken!",
		}
		return c.Status(422).JSON(utils.RequestErr(utils.ERR_INVALID_ENTRY, "Invalid Entry", data))
	}

	// Create User
	db.Create(&user)

	// Create Otp
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	go senders.SendEmail(user, senders.EmailActivate, &otp.Code)

	response := schemas.RegisterResponseSchema{
		ResponseSchema: SuccessResponse("Registration successful"),
		Data:           schemas.EmailRequestSchema{Email: user.Email},
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

	go senders.SendEmail(&user, senders.EmailWelcome, nil)

	return c.Status(200).JSON(SuccessResponse("Account verification successful"))
}

func (endpoint Endpoint) ResendVerificationEmail(c *fiber.Ctx) error {
	db := endpoint.DB
	data := schemas.EmailRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: data.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	if user.IsEmailVerified {
		return c.Status(200).JSON(SuccessResponse("Email already verified"))
	}

	// Send Email
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	go senders.SendEmail(&user, senders.EmailActivate, &otp.Code)

	return c.Status(200).JSON(SuccessResponse("Verification email sent"))
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
	db := endpoint.DB
	user := RequestUser(c)
	reqData := schemas.RefreshTokenRequestSchema{}

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

	user.Access = &access
	user.Refresh = &refresh
	db.Save(&user)

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: SuccessResponse("Tokens refresh successful"),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) SendPasswordResetOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	data := schemas.EmailRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: data.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	// Create Otp
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	go senders.SendEmail(&user, senders.EmailResetPassword, &otp.Code)

	return c.Status(200).JSON(SuccessResponse("Password otp sent"))
}

func (endpoint Endpoint) SetNewPassword(c *fiber.Ctx) error {
	db := endpoint.DB
	data := schemas.SetNewPasswordSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: data.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_EMAIL, "Incorrect Email"))
	}

	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != data.Otp {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	if otp.CheckExpiration() {
		return c.Status(400).JSON(utils.RequestErr(utils.ERR_EXPIRED_OTP, "Expired Otp"))
	}

	// Update Users Password & Delete Otp
	db.Model(&user).Updates(map[string]interface{}{"Password": data.Password})
	db.Delete(&otp)

	go senders.SendEmail(&user, senders.EmailResetPasswordSuccess, nil)

	return c.Status(200).JSON(SuccessResponse("Password reset successful"))
}

func (endpoint Endpoint) SendLoginOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	data := schemas.EmailRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: data.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	if !user.IsEmailVerified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Create Otp
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	go senders.SendEmail(&user, senders.EmailOtpLogin, &otp.Code)

	return c.Status(200).JSON(SuccessResponse("Login otp has been sent"))
}

func (endpoint Endpoint) LoginWithOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	code, err := strconv.ParseUint(c.Params("otp"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(err)
	}

	otp := models.Otp{Code: uint32(code)}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != uint32(code) {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	if otp.CheckExpiration() {
		return c.Status(400).JSON(utils.RequestErr(utils.ERR_EXPIRED_OTP, "Expired Otp"))
	}

	user := models.User{ID: otp.UserId}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INVALID_OWNER, "User Not Found"))
	}

	db.Delete(&otp)

	// Create Auth Tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	user.Access = &access
	user.Refresh = &refresh
	db.Save(&user)

	// Set the access token and refresh token cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: SuccessResponse("Logged in successfully"),
		Data:           schemas.TokensResponseSchema{User: &user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}
