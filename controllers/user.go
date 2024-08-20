package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
