package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/middlewares"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	userRouter := app.Group("/api/users")

	// User AUTHENTICATION
	userRouter.Post("/signup", controllers.SignUp)
	userRouter.Post("/login", middlewares.RateLimiter(), controllers.LogIn)
	userRouter.Get("/logout", controllers.LogOut)
	userRouter.Post("/verify-account", controllers.VerifyAccount)
	userRouter.Get("/validate", controllers.Validate)

	// Password RESET and UPDATE for UNAUTHORIZED users
	userRouter.Post("/forgot-password", middlewares.RateLimiter(), controllers.ForgotPassword)
	userRouter.Post("/verify-password-reset-token", controllers.VerifyPasswordResetToken)
	userRouter.Post("/reset-forgotten-password", controllers.ResetUserPassword)

	// Profile UPDATE for AUTHORIZED users
	userRouter.Patch("/update-my-password", middlewares.Protect(), controllers.UpdateUserPassword)
	userRouter.Patch("/update-me", middlewares.Protect(), controllers.UpdateMe)
	userRouter.Delete("/deactivate-me", middlewares.Protect(), controllers.DeleteMe)
	userRouter.Get("/request-email-change-verification-code",
		middlewares.Protect(),
		controllers.GenerateUserEmailChangeVerificationToken)
	userRouter.Patch("/update-my-email", middlewares.Protect(), controllers.UpdateUserEmail)

	// CURRENT USER PHOTO UPDATE
	userRouter.Post("/file-upload", middlewares.Protect(), controllers.UploadUserPhoto)
	userRouter.Get("/file-delete", middlewares.Protect(), controllers.DeleteUserPhoto)

	// Get CURRENT AUTHORIZED user
	userRouter.Get("/me", controllers.GetCurrentUser)

	userRouter.Get("/:id", controllers.GetUserByParamsID)

	userRouter.Get("/",
		middlewares.Protect(),
		middlewares.RestrictTo(models.AdminRole),
		controllers.GetUsers)
}
