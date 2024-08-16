package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// GetUsers retrieves all users
func GetUsers(c *fiber.Ctx) error {
	var users []models.User
	database.DB.Find(&users)
	return c.JSON(users)
}

// CreateUser creates a new user
func CreateUser(c *fiber.Ctx) error {
	// Parse the request body into the User struct
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		// Log the error and return a bad request response
		log.Printf("Error parsing request body: %v", err)
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	// Log the request body
	fmt.Printf("Received user: %+v\n", user)

	// Create the user in the database
	if err := database.DB.Create(&user).Error; err != nil {
		// Log the error and return an internal server error response
		log.Printf("Error creating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not create user")
	}

	return c.Status(201).JSON(user)
}

// DeleteUser deletes a user by ID
func DeleteUser(c *fiber.Ctx) error {
	// Get the ID from the request URL
	id := c.Params("id")

	// Find the user by ID
	var user models.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If no user is found, return a 404 response
			return c.Status(fiber.StatusNotFound).SendString("User not found")
		}
		// Log any other errors and return a 500 response
		log.Printf("Error finding user: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not find user")
	}

	// Delete the user (soft delete)
	if err := database.DB.Delete(&user).Error; err != nil {
		// Log the error and return a 500 response
		log.Printf("Error deleting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not delete user")
	}

	// Return a success message
	return c.SendString("User successfully deleted")
}

// CreateToken generates a JWT token for the given user ID
func CreateToken(userID string, secret string, expiresIn string) (string, error) {
	// Parse the expiration duration
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", err
	}

	// Define claims
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// LoginUser handles the login logic
func LoginUser(c *fiber.Ctx) error {
	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1) Check if email and password exist
	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Please provide email and password"})
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
	accessToken, err := CreateToken(user.ID.String(), "your_jwt_secret", "24h")
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
