package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/validate"
	"github.com/gofiber/fiber/v2"
)

func (endpoint Endpoint) Login(c *fiber.Ctx) error {
	input, err := validate.ParseLoginInput(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := controllers.GetUserByEmail(database.DB, input.Email)
	if user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}
	if !user.ComparePassword(input.Password) {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_CREDENTIALS, "Invalid Credentials"))
	}

	access := authentication.GenerateAccessToken(user.ID)

	response := schemas.LoginResponseSchema{
		ResponseSchema: schemas.ResponseSchema{Message: "Login successful"}.Init(),
		Data:           schemas.TokensResponseSchema{User: user, Access: access},
	}
	return c.Status(201).JSON(response)
}
