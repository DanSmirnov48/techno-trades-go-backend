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
	authRouter.Get("/logout", endpoint.Logout)
	authRouter.Post("/verify-account", endpoint.VerifyAccount)
	authRouter.Post("/resend-verification-email", endpoint.ResendVerificationEmail)
	authRouter.Get("/validate", endpoint.ValidateMe)
	authRouter.Post("/refresh", midw.AuthMiddleware, endpoint.Refresh)
	authRouter.Post("/forgot-password", midw.RateLimiter, endpoint.SendPasswordResetOtp)
	authRouter.Post("/set-new-password", endpoint.SetNewPassword)
	authRouter.Get("/send-login-otp", endpoint.SendLoginOtp)
	authRouter.Post("/login/:otp", endpoint.LoginWithOtp)

	// Users profile routes (5) for AUTHORIZED users
	users := api.Group("/users")
	users.Patch("/update-my-password", midw.AuthMiddleware, endpoint.UpdateSignedInUserPassword)
	users.Patch("/update-me", midw.AuthMiddleware, endpoint.UpdateMe)
	users.Delete("/deactivate-me", midw.AuthMiddleware, endpoint.DeleteMe)
	users.Get("/send-email-change-otp", midw.AuthMiddleware, endpoint.SendUserEmailChangeOtp)
	users.Patch("/update-my-email", midw.AuthMiddleware, endpoint.UpdateUserEmail)
	users.Get("/:id", endpoint.GetUserByParamsID)
	users.Get("/", midw.AuthMiddleware, midw.Admin, endpoint.GetAllUsers)

	// ### -----------------------PRODUCTS-----------------------
	// Product Routes (8)
	products := api.Group("/products")
	products.Get("/:slug", endpoint.FindProductBySlug)
	products.Get("/:id", endpoint.FindProductById)
	products.Get("/", endpoint.GetAllProducts)

	admin_products := api.Group("/products", midw.AuthMiddleware, midw.Admin)
	admin_products.Post("/new", endpoint.CreateNewProduct)
	admin_products.Patch("/:id/update", endpoint.UpdateProductDetails)
	admin_products.Delete("/:id/delete", endpoint.DeleteProduct)
	admin_products.Patch("/:id/update-discount", endpoint.SetProductDiscount)
	admin_products.Patch("/:id/update-stock", endpoint.UpdateProductStock)

	// ### -----------------------REVIEWS-----------------------
	// Reviews Routes (1)
	reviews := api.Group("/reviews")
	reviews.Post("/product/:id/new", midw.AuthMiddleware, endpoint.CreateNewReview)
}
