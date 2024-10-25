package managers

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
)

// ----------------------------------
// PRODUCT MANAGEMENT
// --------------------------------
type ReviewManager struct{}

func (obj ReviewManager) Create(db *gorm.DB, data schemas.CreateReview, userId, productId *uuid.UUID) (*models.Review, *int, *utils.ErrorResponse) {
	var existingReview models.Review
	if err := db.Where("user_id = ? AND product_id = ?", userId, productId).First(&existingReview).Error; err == nil {
		statusCode := 400
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "User has already reviewed this product")
		return nil, &statusCode, &errData
	}

	review := models.Review{
		Title:     data.Title,
		Comment:   data.Comment,
		Rating:    data.Rating,
		UserId:    *userId,
		ProductId: *productId,
	}

	// Save the review to the database
	if err := db.Create(&review).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to create review")
		return nil, &statusCode, &errData
	}

	// Update product's rating using ProductManager
	productManager := ProductManager{}
	if _, errCode, errData := productManager.UpdateRating(db, productId); errCode != nil {
		return nil, errCode, errData
	}

	// Retrieve the newly created review with User and Product populated
	if err := db.Preload(clause.Associations).Take(&review).Error; err != nil {
		statusCode := 500
		errData := utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Failed to retrieve review")
		return nil, &statusCode, &errData
	}

	return &review, nil, nil
}
