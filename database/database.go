package database

import (
	"fmt"
	"log"
	"os"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	dsn := os.Getenv("POSTGRES_DB")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection established")

	// Auto migrate models
	err = DB.AutoMigrate(&models.User{}, &models.Product{}, &models.Image{})
	if err != nil {
		log.Fatal("failed to migrate database: %w", err)
	}
}
