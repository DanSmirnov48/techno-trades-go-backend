package tests

import (
	"fmt"
	"testing"

	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func login(t *testing.T, app *fiber.App, db *gorm.DB, baseUrl string) {
	t.Run("Login", func(t *testing.T) {
		user := CreateTestUser(db)

		url := fmt.Sprintf("%s/login", baseUrl)
		loginData := schemas.LoginSchema{
			Email:    "invalid@example.com", // Invalid email
			Password: "invalidpassword",
		}

		res := ProcessTestBody(t, app, url, "POST", loginData)

		// # Test for invalid credentials
		// Assert Status code
		assert.Equal(t, 401, res.StatusCode)
		// Parse and assert body
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "failure", body["status"])
		assert.Equal(t, utils.ERR_INVALID_CREDENTIALS, body["code"])
		assert.Equal(t, "Invalid Credentials", body["message"])

		// Test for unverified credentials (email)
		loginData.Email = user.Email
		loginData.Password = "testpassword"
		res = ProcessTestBody(t, app, url, "POST", loginData)
		// Assert Status code
		assert.Equal(t, 401, res.StatusCode)
		// Parse and assert body
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "failure", body["status"])
		assert.Equal(t, utils.ERR_UNVERIFIED_USER, body["code"])
		assert.Equal(t, "Verify your email first", body["message"])

		// Test for valid credentials and verified email address
		userManager.SetAccountVerified(db, user)
		res = ProcessTestBody(t, app, url, "POST", loginData)
		// Assert response
		assert.Equal(t, 201, res.StatusCode)
		// Parse and assert body
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Login successful", body["message"])
	})
}

func TestAuth(t *testing.T) {
	app := fiber.New()
	db := Setup(t, app)
	BASEURL := "/api/v1/auth"

	// Run Auth Endpoint Tests
	login(t, app, db, BASEURL)

	// Drop Tables and Close Connectiom
	DropData(db)
	CloseTestDatabase(db)
}
