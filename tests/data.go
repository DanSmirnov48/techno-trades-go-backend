package tests

import (
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	productManager = managers.ProductManager{}
)

// AUTH
func CreateTestUser(db *gorm.DB) models.User {
	user := models.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "testuser@example.com",
		Password:  "testpassword",
	}
	db.FirstOrCreate(&user, models.User{Email: user.Email})

	user.IsEmailVerified = false
	db.Save(&user)

	return user
}

func CreateTestVerifiedUser(db *gorm.DB) models.User {
	user := models.User{
		FirstName:       "Test",
		LastName:        "Verified",
		Email:           "testverifieduser@example.com",
		Password:        "testpassword",
		IsEmailVerified: true,
	}
	db.FirstOrCreate(&user, models.User{Email: user.Email})
	return user
}

func CreateVerifiedTestAdminUser(db *gorm.DB) models.User {
	user := models.User{
		FirstName:       "Test",
		LastName:        "Verified",
		Email:           "testverifieduser@example.com",
		Password:        "testpassword",
		IsEmailVerified: true,
		Role:            models.AdminRole,
	}
	db.FirstOrCreate(&user, models.User{Email: user.Email})
	return user
}

// PRODUCTS
func CreateNewProduct(db *gorm.DB, userId uuid.UUID) *models.Product {
	rndName := fmt.Sprintf("test_product_%s", utils.GetRandomString(10))
	productData := schemas.CreateProduct{
		Name:         rndName,
		Brand:        "test_products",
		Category:     "test_products",
		Description:  "this is a product description blah blah blah",
		Price:        100,
		CountInStock: 100,
		IsDiscounted: false,
	}
	newProduct := productManager.Create(db, productData, userId)
	return newProduct
}
