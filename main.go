package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/swagger"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	_ "github.com/DanSmirnov48/techno-trades-go-backend/docs"
	"github.com/DanSmirnov48/techno-trades-go-backend/routes"
)

// @title Your API Title
// @version 1.0
// @description This is a sample server.
// @host localhost:8000
// @BasePath /api/v1
func main() {
	cfg := config.GetConfig()
	db := database.ConnectDb(cfg)
	sqlDb, _ := db.DB()

	app := fiber.New()

	app.Use(helmet.New())

	// CORS config
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Access-Control-Allow-Origin, Content-Disposition",
		AllowCredentials: true,
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Options("*", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "hello there"})
	})

	// Set up routes
	routes.SetupRoutes(app, db)
	defer sqlDb.Close()
	log.Fatal(app.Listen(":8000"))
}
