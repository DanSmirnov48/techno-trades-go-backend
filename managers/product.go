package managers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
)

// ----------------------------------
// USER MANAGEMENT
// --------------------------------
type ProductManager struct{}

func (obj ProductManager) Create(db *gorm.DB, productSchema schemas.CreateProduct, userId uuid.UUID) (*models.Product, *fiber.Error) {
	product := models.Product{
		ID:              uuid.New(),
		Name:            productSchema.Name,
		Slug:            slug.Make(productSchema.Name),
		Brand:           productSchema.Brand,
		Category:        productSchema.Category,
		Description:     productSchema.Description,
		Rating:          0,
		Price:           productSchema.Price,
		CountInStock:    productSchema.CountInStock,
		IsDiscounted:    productSchema.IsDiscounted,
		DiscountedPrice: productSchema.DiscountedPrice,
		UserID:          userId,
	}

	if err := db.Create(&product).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Could not create user")
	}

	return &product, nil
}

func (obj ProductManager) DropData(db *gorm.DB) error {
	// Use the GORM Migrator to drop the User table
	if err := db.Migrator().DropTable(&models.Product{}); err != nil {
		return fmt.Errorf("failed to drop user table: %w", err)
	}
	log.Println("Products table dropped successfully.")
	return nil
}
