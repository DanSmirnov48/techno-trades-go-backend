package tests

import (
	"fmt"
	"testing"

	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func login(t *testing.T, app *fiber.App, baseUrl string) {
	t.Run("Login with invalid credentials", func(t *testing.T) {
		url := fmt.Sprintf("%s/login", baseUrl)
		loginData := schemas.LoginSchema{
			Email:    "invalid@example.com",
			Password: "invalidpassword",
		}

		res := ProcessTestBody(t, app, url, "POST", loginData)

		// Test for invalid credentials
		// Assert Status code
		assert.Equal(t, 401, res.StatusCode)

		// Parse and assert body
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "Incorrect email or password", body["error"])
	})
}

func TestAuth(t *testing.T) {
	app := fiber.New()
	db := Setup(t, app)
	BASEURL := "/api/v1/auth"

	// Run Auth Endpoint Tests
	login(t, app, BASEURL)

	// Close Connection
	CloseTestDatabase(db)
}
