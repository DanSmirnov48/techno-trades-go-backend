package middlewares

import (
	"os"

	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
)

// Middleware to attach the user object to the context
func Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the Authorization token from Header
		tokenString, err := utils.GetAuthorizationHeader(c)
		if err != nil {
			fiberErr := err.(*fiber.Error)
			return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
		}

		// Parse the JWT token
		secret := os.Getenv("JWT_SECRET")
		token, err := utils.ParseJWT(tokenString, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired token",
			})
		}

		// Validate the JWT claims
		userID, err := utils.ValidateJWTClaims(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		user, err := controllers.GetUserByID(database.DB, userID)

		// Attach the user object to the context
		c.Locals("user", user)

		// Proceed to the next handler
		return c.Next()
	}
}

// RestrictTo checks if the authenticated user has one of the required roles
func RestrictTo(roles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the user object from the context (set by the Protect middleware)
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "User information is missing. You do not have permission to perform this action.",
			})
		}

		// Check if the user's role is in the allowed roles
		hasPermission := false
		for _, role := range roles {
			if user.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "You do not have permission to perform this action.",
			})
		}

		// User has the required role, proceed to the next handler
		return c.Next()
	}
}
