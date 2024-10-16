package routes

import (
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (endpoint Endpoint) GetAllUsers(c *fiber.Ctx) error {
	db := endpoint.DB

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil || user.Role != models.AdminRole {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	users, _ := userManager.GetAll(db)
	if users == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "Users not found"))
	}

	response := schemas.FindAllUsersResponseSchem{
		ResponseSchema: schemas.ResponseSchema{Message: "All Users Found"}.Init(),
		Data:           schemas.UsersResponseSchem{Users: users, Length: len(users)},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) GetUserByParamsID(c *fiber.Ctx) error {
	db := endpoint.DB

	userId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	user, _ := userManager.GetById(db, *userId)
	if user == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "User not found"))
	}

	response := schemas.FindUserByIdResponseSchem{
		ResponseSchema: schemas.ResponseSchema{Message: "User Found"}.Init(),
		Data:           schemas.UserResponseSchem{Users: user},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) SendForgotPasswordOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	emailSchema := schemas.EmailRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &emailSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
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
	if errCode, errData := ValidateRequest(c, &otpSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
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
	if errCode, errData := ValidateRequest(c, &resetSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
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

func (endpoint Endpoint) UpdateSignedInUserPassword(c *fiber.Ctx) error {
	db := endpoint.DB
	passwordSchema := schemas.UpdateUserPasswordRequestSchema{}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	// Validate request
	if errCode, errData := ValidateRequest(c, &passwordSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if !user.ComparePassword(passwordSchema.CurrentPassword) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Current password is incorrect"))
	}

	if err := db.Model(&user).Updates(map[string]interface{}{
		"Password": passwordSchema.NewPassword,
	}).Error; err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to update password"))
	}

	response := schemas.ResponseSchema{Message: "Password updated successfully"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) SendUserEmailChangeVerificationToken(c *fiber.Ctx) error {
	db := endpoint.DB

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	if !user.Verified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	token, err := user.CreateEmailUpdateVerificationToken()
	if err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to send email update email"))
	}

	db.Save(&user)

	response := schemas.SendPasswordResetOtpResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Email Update Token sent successful"}.Init(),
		Data:           schemas.PasswordResetOtpResponseSchema{Email: user.Email, Otp: token},
	}
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) UpdateUserEmail(c *fiber.Ctx) error {
	db := endpoint.DB
	emailSchema := schemas.UpdateUserEmailRequestSchema{}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	if !user.Verified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Validate request
	if errCode, errData := ValidateRequest(c, &emailSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if user.EmailUpdateVerificationToken != emailSchema.Otp {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Invalid verification code"))
	}

	if err := db.Model(&user).Updates(map[string]interface{}{
		"email":                           emailSchema.NewEmail,
		"email_update_verification_token": nil,
	}).Error; err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to update Email"))
	}

	response := schemas.ResponseSchema{Message: "Email updated successfully"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) DeleteMe(c *fiber.Ctx) error {
	db := endpoint.DB

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	if err := db.Delete(&user).Error; err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Could not delete user"))
	}

	response := schemas.ResponseSchema{Message: "User deleted successfully"}.Init()
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) UpdateMe(c *fiber.Ctx) error {
	db := endpoint.DB
	updateMeSchema := schemas.UpdateUserRequestSchema{}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	// Validate request
	if errCode, errData := ValidateRequest(c, &updateMeSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if err := db.Model(&user).
		Clauses(clause.Returning{}).
		Updates(updateMeSchema).Error; err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to update user information"))
	}

	response := schemas.ResponseSchema{Message: "User updated successfully"}.Init()
	return c.Status(200).JSON(response)
}
