package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
	Age  int
}

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
	result := db.Find(&users)
	return users, result.Error
}
