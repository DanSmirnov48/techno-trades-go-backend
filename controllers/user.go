package controllers

import (
	"log"
	"strings"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/validate"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUsers retrieves all users
func GetUsers(c *fiber.Ctx) error {
	var users []models.User
	database.DB.Find(&users)
	return c.JSON(users)
}

func GetUserByID(db *gorm.DB, userID string) (*models.User, *fiber.Error) {
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	return &user, nil
}

// CreateUser creates a new user
func CreateUser(c *fiber.Ctx) error {
	// Parse and validate the input
	input, err := validate.ParseCreateUserInput(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Create the user in the database
	user := models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  input.Password,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email is not available"})
		}

		// Log the error and return a generic internal server error response
		log.Printf("Error creating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Could not create user")
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// DeleteUser deletes a user by ID
func DeleteUser(c *fiber.Ctx) error {
	// Get the ID from the request URL
	id := c.Params("id")

	user, err := GetUserByID(database.DB, id)
	if err != nil {
		return err
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

// UpdateMe allows the current authenticated user to update their information.
func UpdateMe(c *fiber.Ctx) error {
	// Parse the request body into a map
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get the authenticated user from the context
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User information is missing. You do not have permission to perform this action.",
		})
	}

	// Filter the allowed fields from the request body
	allowedFields := []string{"FirstName", "LastName"}
	filteredBody := utils.FilteredFields(body, allowedFields...)

	// Update the user with the filtered fields and use Returning clause to get the updated data
	if err := database.DB.Model(&user).
		Clauses(clause.Returning{}).
		Updates(filteredBody).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update user information",
		})
	}

	// Return the updated user data directly from the Returning clause
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": filteredBody,
		},
	})
}
