package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/joho/godotenv"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/middlewares"
	"github.com/DanSmirnov48/techno-trades-go-backend/routes"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/mail"
)

func main() {
	if os.Getenv("ENV") != "production" {
		// Load the .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	data := mail.EmailData{
		Name:            "John Doe",
		ConfirmationURL: "https://github.com/go-gomail/gomail",
	}

	// Send the email
	err := mail.SendEmail("", data)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Styled email sent successfully!")
	}

	// Connect to the database
	database.ConnectDB()

	app := fiber.New()

	app.Use(helmet.New())
	app.Use(middlewares.CorsHandler())

	app.Options("*", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "hello there"})
	})

	// Set up routes
	routes.RegisterUserRoutes(app)

	log.Fatal(app.Listen("0.0.0.0:" + port))
}
