package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
		ResponseSchema: SuccessResponse("All Users Found"),
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
		ResponseSchema: SuccessResponse("User Found"),
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

	if !user.IsEmailVerified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Create Otp
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	response := schemas.SendPasswordResetOtpResponseSchema{
		ResponseSchema: SuccessResponse("Password Reset Token sent successful"),
		Data:           schemas.PasswordResetOtpResponseSchema{Email: user.Email, Otp: otp.Code},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) VerifyForottenPasswordOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	otpSchema := schemas.PasswordResetOtpRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &otpSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	otp := models.Otp{Code: otpSchema.Otp}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != uint32(otpSchema.Otp) {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	if otp.CheckExpiration() {
		return c.Status(400).JSON(utils.RequestErr(utils.ERR_EXPIRED_OTP, "Expired Otp"))
	}
	return c.Status(200).JSON(SuccessResponse("Token verification successful"))
}

func (endpoint Endpoint) ResetUserForgottenPassword(c *fiber.Ctx) error {
	db := endpoint.DB
	resetSchema := schemas.UserPasswordResetRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &resetSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user := models.User{Email: resetSchema.Email}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "User not Found"))
	}

	otp := models.Otp{Code: resetSchema.Otp, UserId: user.ID}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != resetSchema.Otp {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	// Update Users Password & Delete Otp
	db.Model(&user).Updates(map[string]interface{}{"Password": resetSchema.NewPassword})
	db.Delete(&otp)

	return c.Status(200).JSON(SuccessResponse("Password updated successfully"))
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

	// Update Users Password
	db.Model(&user).Updates(map[string]interface{}{"Password": passwordSchema.NewPassword})

	return c.Status(200).JSON(SuccessResponse("Password updated successfully"))
}

func (endpoint Endpoint) SendUserEmailChangeOtp(c *fiber.Ctx) error {
	db := endpoint.DB

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	if !user.IsEmailVerified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Create Otp
	otp := models.Otp{UserId: user.ID}
	db.Take(&otp, otp)
	db.Create(&otp)

	response := schemas.SendPasswordResetOtpResponseSchema{
		ResponseSchema: SuccessResponse("Email Update Token sent successful"),
		Data:           schemas.PasswordResetOtpResponseSchema{Email: user.Email, Otp: otp.Code},
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

	if !user.IsEmailVerified {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNVERIFIED_USER, "Verify your email first"))
	}

	// Validate request
	if errCode, errData := ValidateRequest(c, &emailSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	otp := models.Otp{Code: emailSchema.Otp}
	db.Take(&otp, otp)
	if otp.ID == uuid.Nil || otp.Code != emailSchema.Otp {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_INCORRECT_OTP, "Incorrect Otp"))
	}

	// Update Users Email & Delete Otp
	db.Model(&user).Updates(map[string]interface{}{"email": emailSchema.NewEmail})
	db.Delete(&otp)

	return c.Status(200).JSON(SuccessResponse("Email updated successfully"))
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

	return c.Status(200).JSON(SuccessResponse("User deleted successfully"))
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

	return c.Status(200).JSON(SuccessResponse("User updated successfully"))
}
