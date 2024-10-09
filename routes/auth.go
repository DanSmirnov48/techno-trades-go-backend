package routes

import (
	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	validator = utils.Validator()
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

	// Generate access token
	access := auth.GenerateAccessToken(user.ID)

	// Set the access token in a cookie
	auth.SetAuthCookie(c, "accessToken", access, 60)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Login successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) Logout(c *fiber.Ctx) error {
	// Remove the access token cookie
	auth.RemoveAuthCookie(c, "accessToken")

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
