package routes

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterProductRoutes(app *fiber.App) {
	userRouter := app.Group("/api/products")

	// User AUTHENTICATION
	userRouter.Post("/new", controllers.CreateSampleProduct)
}
