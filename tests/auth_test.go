package tests

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
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
		user.IsEmailVerified = true
		db.Save(&user)
		res = ProcessTestBody(t, app, url, "POST", loginData)
		// Assert response
		assert.Equal(t, 201, res.StatusCode)
		// Parse and assert body
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Login successful", body["message"])
	})
}

func register(t *testing.T, app *fiber.App, baseUrl string) {
	t.Run("Register User", func(t *testing.T) {
		url := fmt.Sprintf("%s/register", baseUrl)
		validEmail := "testregisteruser@email.com"
		userData := schemas.RegisterUser{
			FirstName: "TestRegister",
			LastName:  "User",
			Email:     validEmail,
			Password:  "testregisteruserpassword",
		}

		res := ProcessTestBody(t, app, url, "POST", userData)

		// Assert Status code
		assert.Equal(t, 201, res.StatusCode)

		// Parse and assert body
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Registration successful", body["message"])
		expectedData := make(map[string]interface{})
		expectedData["email"] = validEmail
		assert.Equal(t, expectedData, body["data"].(map[string]interface{}))

		// // Verify that a user with the same email cannot be registered again
		res = ProcessTestBody(t, app, url, "POST", userData)
		assert.Equal(t, 422, res.StatusCode)

		// Parse and assert body
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "failure", body["status"])
		assert.Equal(t, utils.ERR_INVALID_ENTRY, body["code"])
		assert.Equal(t, "Invalid Entry", body["message"])
		expectedData = make(map[string]interface{})
		expectedData["email"] = "Email already registered!"
		assert.Equal(t, expectedData, body["data"].(map[string]interface{}))
	})
}

func verifyAccount(t *testing.T, app *fiber.App, db *gorm.DB, baseUrl string) {
	t.Run("Verify Account", func(t *testing.T) {
		user := CreateTestUser(db)
		otp := uint32(1111)

		url := fmt.Sprintf("%s/verify-account", baseUrl)
		emailOtpData := schemas.VerifyAccountRequestSchema{
			EmailRequestSchema: schemas.EmailRequestSchema{Email: user.Email},
			Otp:                otp,
		}

		res := ProcessTestBody(t, app, url, "POST", emailOtpData)

		// Verify that the email verification fails with an invalid otp
		// Assert Status code
		assert.Equal(t, 404, res.StatusCode)

		// Parse and assert body
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "failure", body["status"])
		assert.Equal(t, utils.ERR_INCORRECT_OTP, body["code"])
		assert.Equal(t, "Incorrect Otp", body["message"])

		// Verify that the email verification succeeds with a valid otp
		realOtp := models.Otp{UserId: user.ID}
		db.Take(&realOtp, realOtp)
		db.Save(&realOtp) // Create or save
		emailOtpData.Otp = realOtp.Code
		res = ProcessTestBody(t, app, url, "POST", emailOtpData)
		assert.Equal(t, 200, res.StatusCode)

		// Parse and assert body
		body = ParseResponseBody(t, res.Body).(map[string]interface{})
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Account verification successful", body["message"])
	})
}

func logout(t *testing.T, app *fiber.App, baseUrl string) {
	t.Run("Logout", func(t *testing.T) {
		url := fmt.Sprintf("%s/logout", baseUrl)
		req := httptest.NewRequest("GET", url, nil)
		res, _ := app.Test(req)

		// Parse and assert body
		body := ParseResponseBody(t, res.Body).(map[string]interface{})
		// Assert Status code
		assert.Equal(t, 200, res.StatusCode)
		assert.Equal(t, "success", body["status"])
		assert.Equal(t, "Logout successful", body["message"])
	})
}

func TestAuth(t *testing.T) {
	app := fiber.New()
	db := Setup(t, app)
	BASEURL := "/api/v1/auth"

	// Run Auth Endpoint Tests
	register(t, app, BASEURL)
	verifyAccount(t, app, db, BASEURL)
	login(t, app, db, BASEURL)
	logout(t, app, BASEURL)

	// Drop Tables and Close Connectiom
	database.DropTables(db)
	CloseTestDatabase(db)
}
