package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func CorsHandler() fiber.Handler {
	return cors.New(cors.Config{
		Next:             nil,
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept",
		AllowCredentials: true,
		ExposeHeaders:    "",
		MaxAge:           0,
	})
}

// RateLimiter is a middleware that applies rate limiting to routes.
func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		// Limit the maximum number of requests per period
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // Per 1 minute
		// Customize the response when limit is reached
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
		},
	})
}
