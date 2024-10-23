package managers

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
)

// ----------------------------------
// PRODUCT MANAGEMENT
// --------------------------------
type ProductManager struct{}

func (obj ProductManager) Create(db *gorm.DB, data schemas.CreateProduct, userId uuid.UUID) *models.Product {
	product := utils.ConvertStructData(data, models.Product{}).(*models.Product)
	product.ID = uuid.New()
	product.Slug = slug.Make(product.Name)
	product.Rating = 0
	product.UserID = userId

	db.Create(&product)

	return product
}

func (obj ProductManager) DropData(db *gorm.DB) error {
	// Use the GORM Migrator to drop the User table
	if err := db.Migrator().DropTable(&models.Product{}); err != nil {
		return fmt.Errorf("failed to drop user table: %w", err)
	}
	log.Println("Products table dropped successfully.")
	return nil
}
