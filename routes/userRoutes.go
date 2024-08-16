package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	userRouter := app.Group("/api/users")

	userRouter.Get("/", controllers.GetUsers)
	userRouter.Post("/", controllers.CreateUser)
	userRouter.Delete("/:id", controllers.DeleteUser)

	userRouter.Post("/login", controllers.LoginUser)
	userRouter.Post("/logout", controllers.LogoutUser)
}
