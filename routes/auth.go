package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
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
	access := authentication.GenerateAccessToken(user.ID)

	// Set the access token in a cookie
	authentication.SetAuthCookie(c, "accessToken", access, 60)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Login successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) Logout(c *fiber.Ctx) error {
	// Remove the access token cookie
	authentication.RemoveAuthCookie(c, "accessToken")

	response := schemas.ResponseSchema{Message: "Logout successful"}.Init()
	return c.Status(200).JSON(response)
}
