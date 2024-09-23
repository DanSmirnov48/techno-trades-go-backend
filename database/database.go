package database

import (
	"fmt"
	"log"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	cfg := config.GetConfig()

	dbUrlTemplate := "host=%s port=%s user=%s dbname=%s password=%s"

	dsn := fmt.Sprintf(
		dbUrlTemplate,
		cfg.PostgresServer,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresDB,
		cfg.PostgresPassword,
	)

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
