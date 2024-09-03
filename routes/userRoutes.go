package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/middlewares"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	userRouter := app.Group("/api/users")

	userRouter.Get("/", controllers.GetUsers)
	userRouter.Post("/", controllers.CreateUser)
	userRouter.Delete("/:id", controllers.DeleteUser)

	userRouter.Post("/login", middlewares.RateLimiter(), controllers.LoginUser)
	userRouter.Post("/logout", controllers.LogoutUser)

	userRouter.Get("/me", controllers.GetCurrentUser)
	userRouter.Get("/protected", middlewares.Protect(), controllers.ProtectedEndpoint)

	userRouter.Get("/admin",
		middlewares.Protect(),
		middlewares.RestrictTo(models.AdminRole),
		controllers.AdminRestictedRoute)

	userRouter.Patch("/update-my-password", middlewares.Protect(), controllers.UpdateUserPassword)

	userRouter.Post("/forgot-password", middlewares.RateLimiter(), controllers.ForgotPassword)
	userRouter.Post("/verify-reset-token", controllers.VerifyPasswordResetToken)
	userRouter.Post("/reset-forgotten-password", controllers.ResetUserPassword)

	userRouter.Patch("/update-me", middlewares.Protect(), controllers.UpdateMe)

	userRouter.Post("/file-upload", middlewares.Protect(), controllers.UploadUserPhoto)
	userRouter.Get("/file-delete", middlewares.Protect(), controllers.DeleteUserPhoto)
}
