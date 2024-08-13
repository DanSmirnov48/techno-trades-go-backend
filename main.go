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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		// Call the CreateUser function to create a new user
		user, err := CreateUser(db, userReq.Name, userReq.Age)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		// Return the created user as a JSON response
		return c.JSON(user)
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

func CreateUser(db *gorm.DB, name string, age int) (*User, error) {
	// Create a new User instance
	user := User{Name: name, Age: age}

	// Use GORM's Create method to insert the new user into the database
	if err := db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Return the created user
	return &user, nil
}

// BeforeCreate is a GORM hook that runs before a User is created
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	fmt.Println("Running BeforeCreate function.")
	fmt.Printf("User created: Name=%s, Age=%d", u.Name, u.Age)

	return nil
}

// AfterCreate is a GORM hook that runs after a User is created
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	// Log the details of the created user
	fmt.Println("Running AfterCreate function.")
	log.Printf("User created: ID=%d, Name=%s, Age=%d", u.ID, u.Name, u.Age)

	return nil
}
