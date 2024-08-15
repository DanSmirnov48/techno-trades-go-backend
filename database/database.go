package database

import (
	"fmt"
	"log"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection established")

	// Auto migrate models
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("failed to migrate database: %w", err)
	}
}
