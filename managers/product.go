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

func (obj ProductManager) GetAll(db *gorm.DB) ([]*models.Product, *int, *utils.ErrorResponse) {
	products := []*models.Product{}
	db.Find(&products)
	if len(products) == 0 {
		status_code := 404
		errData := utils.RequestErr(utils.ERR_NON_EXISTENT, "No products found")
		return nil, &status_code, &errData
	}

	return products, nil, nil
}

func (obj ProductManager) GetById(db *gorm.DB, id uuid.UUID) (*models.Product, *int, *utils.ErrorResponse) {
	product := models.Product{ID: id}
	db.Take(&product, product)
	if product.ID == uuid.Nil {
		status_code := 404
		errData := utils.RequestErr(utils.ERR_NON_EXISTENT, "Product does not exist")
		return nil, &status_code, &errData
	}
	return &product, nil, nil
}

func (obj ProductManager) GetBySlug(db *gorm.DB, slug string) (*models.Product, *int, *utils.ErrorResponse) {
	product := models.Product{Slug: slug}
	db.Take(&product, product)
	if product.ID == uuid.Nil {
		status_code := 404
		errData := utils.RequestErr(utils.ERR_NON_EXISTENT, "Product does not exist")
		return nil, &status_code, &errData
	}
	return &product, nil, nil
}

func (obj ProductManager) UpdateDiscount(db *gorm.DB, id uuid.UUID, data schemas.UpdateDiscount) (*models.Product, *int, *utils.ErrorResponse) {
	product := models.Product{ID: id}
	db.Take(&product, product)
	if product.ID == uuid.Nil {
		status_code := 404
		errData := utils.RequestErr(utils.ERR_NON_EXISTENT, "Product does not exist")
		return nil, &status_code, &errData
	}

	if data.DiscountedPrice >= product.Price {
		statusCode := 400
		errData := utils.RequestErr(utils.ERR_INVALID_ENTRY, "Discounted price has to be lower than original!")
		return nil, &statusCode, &errData
	}

	product.IsDiscounted = data.IsDiscounted
	if data.IsDiscounted {
		product.DiscountedPrice = data.DiscountedPrice
	} else {
		product.DiscountedPrice = 0.0
	}

	if err := db.Save(product).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to update product discount")
		return nil, &statusCode, &errData
	}

	return &product, nil, nil
}

func (pm ProductManager) UpdateStock(db *gorm.DB, id uuid.UUID, stockChange int) (*models.Product, *int, *utils.ErrorResponse) {
	product := models.Product{ID: id}
	db.Take(&product, product)
	if product.ID == uuid.Nil {
		status_code := 404
		errData := utils.RequestErr(utils.ERR_NON_EXISTENT, "Product does not exist")
		return nil, &status_code, &errData
	}

	// Calculate the new stock value
	newStock := product.CountInStock + stockChange

	// Ensure the stock doesn't fall below zero
	if newStock < 0 {
		statusCode := 400
		errData := utils.RequestErr(utils.ERR_INVALID_ENTRY, "Insufficient stock to complete the operation")
		return nil, &statusCode, &errData
	}

	// Update product stock
	product.CountInStock = newStock

	// Save the updated product in the database
	if err := db.Save(product).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to update product stock")
		return nil, &statusCode, &errData
	}

	return &product, nil, nil
}

func (obj ProductManager) UpdateRating(db *gorm.DB, productId *uuid.UUID) (*models.Product, *int, *utils.ErrorResponse) {
	// Fetch all reviews for the product
	var reviews []models.Review
	if err := db.Where("product_id = ?", productId).Find(&reviews).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to fetch reviews for product")
		return nil, &statusCode, &errData
	}

	// Calculate the average rating
	var totalRating int
	for _, review := range reviews {
		totalRating += review.Rating
	}

	// If there are reviews, calculate the average
	var averageRating float64
	if len(reviews) > 0 {
		averageRating = float64(totalRating) / float64(len(reviews))
	}

	// Update the product with the new average rating
	var product models.Product
	if err := db.Model(&product).Where("id = ?", productId).Update("rating", averageRating).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to update product rating")
		return nil, &statusCode, &errData
	}

	// Fetch the updated product
	if err := db.Take(&product, productId).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to retrieve updated product")
		return nil, &statusCode, &errData
	}

	return &product, nil, nil
}

func (obj ProductManager) DropData(db *gorm.DB) error {
	// Use the GORM Migrator to drop the User table
	if err := db.Migrator().DropTable(&models.Product{}); err != nil {
		return fmt.Errorf("failed to drop user table: %w", err)
	}
	log.Println("Products table dropped successfully.")
	return nil
}
