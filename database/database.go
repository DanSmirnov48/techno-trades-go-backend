package database

import (
	"fmt"
	"log"
	"os"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Models() []interface{} {
	return []interface{}{
		&models.User{},
		&models.Product{},
		&models.Image{},
		&models.Otp{},
		&models.Review{},
	}
}

func MakeMigrations(db *gorm.DB) {
	models := Models()
	for _, model := range models {
		db.AutoMigrate(model)
	}
}

func CreateTables(db *gorm.DB) {
	models := Models()
	for _, model := range models {
		db.Migrator().CreateTable(model)
	}
}

func DropTables(db *gorm.DB) {
	models := Models()
	for _, model := range models {
		db.Migrator().DropTable(model)
	}
}

func ConnectDb(cfg config.Config, logs ...bool) *gorm.DB {
	dbUrlTemplate := "host=%s port=%s user=%s dbname=%s password=%s"
	dsn := fmt.Sprintf(
		dbUrlTemplate,
		cfg.PostgresServer,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresDB,
		cfg.PostgresPassword,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to the database! \n", err.Error())
		os.Exit(2)
	}
	log.Println("Connected to the database successfully")

	if len(logs) == 0 {
		// When extra parameter is passed, don't do the following (from sockets)
		log.Println("Running Migrations")

		// Add UUID extension
		result := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
		if result.Error != nil {
			log.Fatal("failed to create extension: " + result.Error.Error())
		}
		// Add Migrations
		MakeMigrations(db)
	}
	return db
}
