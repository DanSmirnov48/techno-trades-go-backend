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
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
)

func main() {
	if os.Getenv("ENV") != "production" {
		// Load the .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
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

	app.Post("/file-upload", UploadAvatar)

	// Set up routes
	routes.SetupRoutes(app)

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func UploadAvatar(c *fiber.Ctx) error {
	// Retrieve the file from the form data
	file, err := c.FormFile("upload")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get the file",
		})
	}

	// Initialize S3 client
	s3Client, err := utils.NewS3Client()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to initialize S3 client",
		})
	}

	// Upload the file to S3
	fileURL, err := s3Client.UploadFile(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to upload file to S3",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": fileURL,
		},
	})
}
