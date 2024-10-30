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

	var users []*models.User
	db.Find(&users)
	if users == nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "Users not found"))
	}

	response := schemas.ManyUsersResponseSchem{
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

	user := models.User{ID: *userId}
	db.Take(&user, user)
	if user.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "User not found"))
	}

	response := schemas.SingleUserResponseSchem{
		ResponseSchema: SuccessResponse("User Found"),
		Data:           schemas.UserResponseSchem{Users: &user},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) UpdateSignedInUserPassword(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	passwordSchema := schemas.UpdateUserPasswordRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &passwordSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if !utils.CheckPasswordHash(passwordSchema.CurrentPassword, user.Password) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Current password is incorrect"))
	}

	// Update Users Password
	db.Model(&user).Updates(map[string]interface{}{"Password": passwordSchema.NewPassword})

	response := schemas.SingleUserResponseSchem{
		ResponseSchema: SuccessResponse("Password updated successfully"),
		Data:           schemas.UserResponseSchem{Users: user},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) SendUserEmailChangeOtp(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)

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
	user := RequestUser(c)
	emailSchema := schemas.UpdateUserEmailRequestSchema{}

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

	response := schemas.SingleUserResponseSchem{
		ResponseSchema: SuccessResponse("Email updated successfully"),
		Data:           schemas.UserResponseSchem{Users: user},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) DeleteMe(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)

	if err := db.Delete(&user).Error; err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Could not delete user"))
	}

	return c.Status(200).JSON(SuccessResponse("User deleted successfully"))
}

func (endpoint Endpoint) UpdateMe(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	updateMeSchema := schemas.UpdateUserRequestSchema{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &updateMeSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if err := db.Model(&user).
		Clauses(clause.Returning{}).
		Updates(updateMeSchema).Error; err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to update user information"))
	}

	response := schemas.SingleUserResponseSchem{
		ResponseSchema: SuccessResponse("User updated successfully"),
		Data:           schemas.UserResponseSchem{Users: user},
	}
	return c.Status(201).JSON(response)
}
