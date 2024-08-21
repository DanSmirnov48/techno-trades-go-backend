package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"
	"github.com/DanSmirnov48/techno-trades-go-backend/middlewares"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	userRouter := app.Group("/api/users")

	userRouter.Get("/", controllers.GetUsers)
	userRouter.Post("/", controllers.CreateUser)
	userRouter.Delete("/:id", controllers.DeleteUser)

	userRouter.Post("/login", middlewares.RateLimiter(), controllers.LoginUser)
	userRouter.Post("/logout", controllers.LogoutUser)

	userRouter.Get("/me", controllers.DecodeJWT)
	userRouter.Get("/protected", middlewares.Protect(), controllers.ProtectedEndpoint)

	userRouter.Get("/admin",
		middlewares.Protect(),
		middlewares.RestrictTo(models.AdminRole),
		controllers.AdminRestictedRoute)
}
