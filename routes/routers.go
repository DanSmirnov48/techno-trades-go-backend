package routes

import (
	midw "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

type Endpoint struct {
	DB *gorm.DB
}

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	midw := midw.Middleware{DB: db}
	endpoint := Endpoint{DB: db}

	api := app.Group("/api/v1")

	// HealthCheck Route (1)
	api.Get("/healthcheck", HealthCheck)

	// Auth Routes (7)
	authRouter := api.Group("/auth")
	authRouter.Post("/register", endpoint.Register)
	authRouter.Post("/login", midw.RateLimiter, endpoint.Login)
	authRouter.Get("/send-magic-link", midw.RateLimiter, endpoint.SendMagicLink)
	authRouter.Post("/login/:token", midw.RateLimiter, endpoint.MagicLinkLogin)
	authRouter.Get("/logout", endpoint.Logout)
	authRouter.Post("/verify-account", endpoint.VerifyAccount)
	authRouter.Get("/validate", endpoint.ValidateMe)
	authRouter.Post("/refresh", midw.AuthMiddleware, endpoint.Refresh)

	// Password RESET Routes (3) for UNAUTHORIZED users
	reset := api.Group("/reset")
	reset.Post("/forgot-password", midw.RateLimiter, endpoint.SendForgotPasswordOtp)
	reset.Post("/verify-password-reset-token", endpoint.VerifyForottenPasswordOtp)
	reset.Post("/reset-forgotten-password", endpoint.ResetUserForgottenPassword)

	// Users profile routes (5) for AUTHORIZED users
	users := api.Group("/users")
	users.Patch("/update-my-password", midw.AuthMiddleware, endpoint.UpdateSignedInUserPassword)
	users.Patch("/update-me", midw.AuthMiddleware, endpoint.UpdateMe)
	users.Delete("/deactivate-me", midw.AuthMiddleware, endpoint.DeleteMe)
	users.Get("/request-email-change-verification-code", midw.AuthMiddleware, endpoint.SendUserEmailChangeOtp)
	users.Patch("/update-my-email", midw.AuthMiddleware, endpoint.UpdateUserEmail)
	users.Get("/:id", endpoint.GetUserByParamsID)
	users.Get("/", midw.AuthMiddleware, midw.Admin, endpoint.GetAllUsers)

	// ### -----------------------PRODUCTS-----------------------
	// Product Routes (1)
	products := api.Group("/products")
	products.Post("/new", midw.AuthMiddleware, midw.Admin, endpoint.CreateNewProduct)
}
