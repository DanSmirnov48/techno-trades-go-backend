package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/joho/godotenv"

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

	// Open the file
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to open the file",
		})
	}
	defer f.Close()

	// Initialize an AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %v", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Buffer to read the file content
	buf := new(bytes.Buffer)
	buf.ReadFrom(f)

	// Define the S3 bucket name and key (path in the bucket)
	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	key := fmt.Sprintf("avatars/%s", filepath.Base(file.Filename))

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		Body:                 bytes.NewReader(buf.Bytes()),
		ContentLength:        aws.Int64(file.Size),
		ContentType:          aws.String(file.Header.Get("Content-Type")),
		ContentDisposition:   aws.String("inline"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to upload file to S3",
		})
	}

	// Return the S3 URL for the uploaded file
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"url": url,
		},
	})
}
