package authentication

import (
	"strings"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Middleware struct {
	DB *gorm.DB
}

func GetUser(token string, db *gorm.DB) (*models.User, *string) {
	if !strings.HasPrefix(token, "Bearer ") {
		err := "Auth Bearer Not Provided"
		return nil, &err
	}
	user, err := DecodeAccessToken(token[7:], db)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (mid Middleware) AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	db := mid.DB

	if len(token) < 1 {
		return c.Status(401).JSON("Unauthorized User!")
	}
	user, err := GetUser(token, db)
	if err != nil {
		return c.Status(401).JSON(*err)
	}
	c.Locals("user", user)
	return c.Next()
}
