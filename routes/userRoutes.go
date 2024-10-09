package routes

import (
	midw "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	c "github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

type Endpoint struct {
	DB *gorm.DB
}

func RegisterUserRoutes(app *fiber.App, db *gorm.DB) {
	midw := midw.Middleware{DB: db}
	endpoint := Endpoint{DB: db}

	api := app.Group("/api/v1")

	// HealthCheck Route (1)
	api.Get("/healthcheck", HealthCheck)

	// Auth Routes (7)
	authRouter := api.Group("/auth")
	authRouter.Post("/signup", endpoint.Register)
	authRouter.Post("/login", midw.RateLimiter, endpoint.Login)
	authRouter.Get("/request-magic-link-login", midw.RateLimiter, c.RequestMagicLink)
	authRouter.Post("/login/:token", midw.RateLimiter, c.LogInWithMagicLink)
	authRouter.Get("/logout", endpoint.Logout)
	authRouter.Post("/verify-account", endpoint.VerifyAccount)
	authRouter.Get("/validate", c.Validate)

	// Password RESET Routes (3) for UNAUTHORIZED users
	reset := api.Group("/reset")
	reset.Post("/forgot-password", midw.RateLimiter, c.ForgotPassword)
	reset.Post("/verify-password-reset-token", c.VerifyPasswordResetToken)
	reset.Post("/reset-forgotten-password", c.ResetUserPassword)

	// Users profile routes (5) for AUTHORIZED users
	users := api.Group("/users")
	users.Patch("/update-my-password", midw.AuthMiddleware, c.UpdateUserPassword)
	users.Patch("/update-me", midw.AuthMiddleware, c.UpdateMe)
	users.Delete("/deactivate-me", midw.AuthMiddleware, c.DeleteMe)
	users.Get("/request-email-change-verification-code", midw.AuthMiddleware, c.GenerateUserEmailChangeVerificationToken)
	users.Patch("/update-my-email", midw.AuthMiddleware, c.UpdateUserEmail)

	// CURRENT USER PHOTO UPDATE
	users.Post("/file-upload", midw.AuthMiddleware, c.UploadUserPhoto)
	users.Get("/file-delete", midw.AuthMiddleware, c.DeleteUserPhoto)

	// Get CURRENT AUTHORIZED user
	users.Get("/me", c.GetCurrentUser)

	users.Get("/:id", c.GetUserByParamsID)

	users.Get("/", midw.AuthMiddleware, midw.RestrictTo(models.AdminRole), c.GetUsers)
}
