package routes

import (
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (endpoint Endpoint) SendForgotPasswordOtp(c *fiber.Ctx) error {
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

	// Generate a password reset token
	token, err := user.CreatePasswordResetVerificationToken()
	if err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to generate password reset code"))
	}

	db.Save(&user)

	response := schemas.SendPasswordResetOtpResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Password Reset Token sent successful"}.Init(),
		Data:           schemas.PasswordResetOtpResponseSchema{Email: user.Email, Otp: token},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) VerifyForottenPasswordResetToken(c *fiber.Ctx) error {
	db := endpoint.DB
	otpSchema := schemas.PasswordResetOtpRequestSchema{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &otpSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(otpSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	var user *models.User

	if err := db.Where("password_reset_token = ? AND password_reset_token_expires > ?", otpSchema.Otp, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, "Refresh token is invalid or expired"))
		}
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Error checking user credentials"))
	}

	response := schemas.ResponseSchema{Message: "Token verification successful"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) ResetUserForgottenPassword(c *fiber.Ctx) error {
	db := endpoint.DB
	resetSchema := schemas.UserPasswordResetRequestSchema{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &resetSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(resetSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	var user *models.User

	if err := db.Where("email = ? AND password_reset_token = ? AND password_reset_token_expires > ?", resetSchema.Email, resetSchema.Otp, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, "Refresh token is invalid or expired"))
		}
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Error checking user credentials"))
	}

	if err := db.Model(&user).Updates(map[string]interface{}{
		"Password":                  resetSchema.NewPassword,
		"PasswordResetToken":        nil,
		"PasswordResetTokenExpires": nil,
	}).Error; err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to update password"))
	}

	response := schemas.ResponseSchema{Message: "Password updated successfully"}.Init()
	return c.Status(200).JSON(response)
}
