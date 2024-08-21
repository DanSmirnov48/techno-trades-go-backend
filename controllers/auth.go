package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/validate"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// LoginUser handles the login logic
func LoginUser(c *fiber.Ctx) error {
	// 1) Parse and validate the login input.
	input, err := validate.ParseLoginInput(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User

	// 2) Check if user exists and retrieve their password
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect email or password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking user credentials"})
	}

	// 3) Compare the provided password with the stored password using the ComparePassword method
	if !user.ComparePassword(input.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect email or password"})
	}

	// 4) Create a JWT token
	accessToken, err := utils.CreateToken(user.ID.String(), "24h")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating token"})
	}

	// 5) Set the token in a cookie
	isSecure := false
	if proto, ok := c.GetReqHeaders()["X-Forwarded-Proto"]; ok {
		for _, p := range proto {
			if strings.ToLower(p) == "https" {
				isSecure = true
				break
			}
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "strict",
	})

	// 6) Remove password from output
	user.Password = "" // Don't send the password back

	// 7) Send response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"token":  accessToken,
		"data": fiber.Map{
			"user": user,
		},
	})
}

// LogoutUser handles the user logout
func LogoutUser(c *fiber.Ctx) error {
	// Clear the access token cookie
	c.ClearCookie("accessToken")

	// Send response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// DecodeJWT verifies the JWT token, extracts the user ID, and retrieves the user from the database.
func DecodeJWT(c *fiber.Ctx) error {
	// Get the JWT from the cookies
	tokenString := c.Cookies("accessToken")

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing or invalid JWT",
		})
	}

	// Parse the JWT token
	secret := os.Getenv("JWT_SECRET")
	token, err := utils.ParseJWT(tokenString, secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate the token claims
	userID, err := utils.ValidateJWTClaims(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Call the GetUserByID function
	user, err := GetUserByID(database.DB, userID)
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Attach the user object to the context
	c.Locals("user", user)

	// Return the user object in the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": user,
		},
	})
}

// Example handler to access the user object
func ProtectedEndpoint(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User information is missing. You do not have permission to perform this action.",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func AdminRestictedRoute(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User information is missing. You do not have permission to perform this action.",
		})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}
