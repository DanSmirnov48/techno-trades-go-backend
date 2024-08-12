package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
	Age  int
}

func main() {
	// Open a connection to the database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{})

	// Define multiple users
	users := []User{
		{Name: "John", Age: 25},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 35},
	}

	// Create multiple users with a single operation
	result := db.Create(&users)

	// Check for errors
	if result.Error != nil {
		panic(result.Error)
	}

	// Print the IDs of the created users
	for _, user := range users {
		fmt.Printf("Created User: ID=%d, Name=%s, Age=%d\n", user.ID, user.Name, user.Age)
	}
}
