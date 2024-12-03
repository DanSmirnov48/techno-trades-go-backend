package authentication

import (
	"strings"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized User!"))
	}
	user, err := GetUser(token, db)
	if err != nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_INVALID_TOKEN, *err))
	}
	c.Locals("user", user)
	return c.Next()
}

func (mid Middleware) RateLimiter(c *fiber.Ctx) error {
	return limiter.New(limiter.Config{
		// Limit the maximum number of requests per period
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // Per 1 minute
		// Customize the response when limit is reached
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(utils.RequestErr(utils.ERR_REQUEST_LIMIT, "Too many requests"))
		},
	})(c)
}

func (mid Middleware) Admin(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil || user.AccountType == models.AccountTypeBuyer {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}
	return c.Next()
}
