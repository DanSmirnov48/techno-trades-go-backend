package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
)

type User = models.User

func main() {
	if os.Getenv("ENV") != "production" {
		// Load the .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	// Call the ConnectDB function with the DSN (in this case, the SQLite database file name)
	db, err := ConnectDB("test.db")
	if err != nil {
		panic(fmt.Sprintf("Error connecting to the database: %v", err))
	}

	// Use the db connection for further operations, e.g., CRUD operations
	fmt.Println("Database connection established successfully")
	// Now you can use the `db` instance for database operations

	app := fiber.New()

	// GET /users - Retrieve all users
	app.Get("/users", func(c *fiber.Ctx) error {
		users, err := GetUsers(db)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(users)
	})

	// POST /users - Create a new user
	app.Post("/users", func(c *fiber.Ctx) error {
		// Define a structure to hold the request body data
		type UserRequest struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		// Parse the request body into the UserRequest struct
		var userReq UserRequest
		if err := c.BodyParser(&userReq); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		// Call the CreateUser function to create a new user
		user, err := CreateUser(db, userReq.Name, userReq.Age)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
		}

		// Return the created user as a JSON response
		return c.JSON(user)
	})

	// DELETE /users/:id - Soft delete a user
	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		// Parse the user ID from the route parameters
		id, err := c.ParamsInt("id")
		if err != nil || id <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		// Call the DeleteUser function to perform a soft delete
		if err := DeleteUser(db, uint(id)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Return a success response
		return c.SendStatus(fiber.StatusNoContent)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "hello there"})
	})

	log.Fatal(app.Listen("0.0.0.0:" + port))

}

// ConnectDB initializes and returns a GORM database connection
func ConnectDB(dsn string) (*gorm.DB, error) {
	// Connect to the database
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate automatically creates tables and updates the schema as needed
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Return the database connection
	return db, nil
}

func GetUsers(db *gorm.DB) ([]User, error) {
	var users []User
	// result := db.Unscoped().Find(&users) FIND ALL RECORDS INCLUDING SOFT DELETED
	result := db.Find(&users)
	return users, result.Error
}

// --------------------------------------------CREATE------------------------------------------
func CreateUser(db *gorm.DB, name string, age int) (*User, error) {
	// Create a new User instance
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john2.doe@test.com",
		Password:  "securepassword",
	}

	// Use GORM's Create method to insert the new user into the database
	if err := db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Return the created user
	return &user, nil
}

// --------------------------------------------DELETE------------------------------------------
// DeleteUser performs a soft delete on the user with the specified ID
func DeleteUser(db *gorm.DB, id uint) error {
	// Use GORM's Delete method to perform a soft delete
	if err := db.Delete(&User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("User with ID=%d has been soft deleted\n", id)

	return nil
}
