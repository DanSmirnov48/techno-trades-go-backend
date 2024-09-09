package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/middlewares"
	"github.com/DanSmirnov48/techno-trades-go-backend/routes"
)

func main() {
	if os.Getenv("ENV") != "production" {
		// Load the .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	sendEmail()

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

func sendEmail() {
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("EMAIL_USERNAME"))
	m.SetHeader("To", "")
	m.SetHeader("Subject", "Test Email from Go")
	m.SetBody("text/html", "<h1>Hello</h1><p>This is a test email!</p>")

	d := gomail.NewDialer(
		os.Getenv("EMAIL_HOST"),
		587,
		os.Getenv("EMAIL_USERNAME"),
		os.Getenv("EMAIL_PASSWORD"),
	)

	if err := d.DialAndSend(m); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Email sent successfully!")
	}
}
