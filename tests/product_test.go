package tests

import (
	"fmt"
	"testing"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func create(t *testing.T, app *fiber.App, db *gorm.DB, baseUrl string) {
	t.Run("Create Product", func(t *testing.T) {
		adminUser := CreateVerifiedTestAdminUser(db)

		loginUrl := fmt.Sprintf("%s/login", "/api/v1/auth")
		loginData := schemas.LoginSchema{
			Email:    adminUser.Email,
			Password: "testpassword",
		}

		// ### Log in as Admin User
		// Use ProcessTestBody to send the login request
		res := ProcessTestBody(t, app, loginUrl, "POST", loginData)
		assert.Equal(t, 201, res.StatusCode)

		// Parse the login response body to extract the access token
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Login successful", body["message"])

		// Extract the access token from the login response data
		tokenData := body["data"].(map[string]interface{})
		accessToken := tokenData["access"].(string)

		// ### Create new Product
		url := fmt.Sprintf("%s/new", baseUrl)
		productData := schemas.CreateProduct{
			Name:         fmt.Sprintf("test_product_%s", utils.GetRandomString(10)),
			Brand:        "test_products",
			Category:     "test_products",
			Description:  "this is a product description blah blah blah",
			Price:        100,
			CountInStock: 100,
			IsDiscounted: false,
		}

		// Send the request to create a new product, including the access token in the Authorization header
		res = ProcessTestBody(t, app, url, "POST", productData, accessToken)
		assert.Equal(t, 201, res.StatusCode)

		// Parse and assert the response body for creating a product
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Product created successfully", body["message"])
	})
}
func TestProduct(t *testing.T) {
	app := fiber.New()
	db := Setup(t, app)
	BASEURL := "/api/v1/products"

	// Run Product Endpoint Tests
	create(t, app, db, BASEURL)

	// Drop Tables and Close Connectiom
	database.DropTables(db)
	CloseTestDatabase(db)
}
