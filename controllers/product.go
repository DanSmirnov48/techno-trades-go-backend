package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
)

func CreateSampleProduct(c *fiber.Ctx) error {
	userID, err := uuid.Parse("014e18d8-dfa6-4065-9974-e318f53aa0ae")
	if err != nil {
		log.Fatalf("Error parsing UserID: %v", err)
	}

	product := models.Product{
		ID:           uuid.New(),
		Name:         "Sample Product",
		Slug:         "sample-product",
		Brand:        "Sample Brand",
		Category:     "Sample Category",
		Description:  "This is a sample product description.",
		Rating:       4.5,
		Price:        29.99,
		CountInStock: 100,
		IsDiscounted: false,
		UserID:       userID,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		// Log the error and return a generic internal server error response
		log.Printf("Error creating product: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not create product")
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}
