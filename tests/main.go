package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDatabase() *gorm.DB {
	var err error
	var DB *gorm.DB
	cfg := config.GetConfig()
	dbUrlTemplate := "host=%s port=%s user=%s dbname=%s password=%s"

	dsn := fmt.Sprintf(
		dbUrlTemplate,
		cfg.PostgresServer,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.TestPostgresDB,
		cfg.PostgresPassword,
	)

	DB, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection established")
	return DB
}

func DropData(db *gorm.DB) {
	if err := db.Migrator().DropTable(&models.User{}); err != nil {
		log.Fatalf("Failed to drop table: %v", err)
	}
	log.Println("Test tables dropped successfully.")
}

func CreateTables(db *gorm.DB) {
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	log.Println("Test tables created successfully.")
}

func CloseTestDatabase(db *gorm.DB) {
	sqlDB, _ := db.DB()

	if err := sqlDB.Close(); err != nil {
		log.Fatalf("Failed to close database connection: %v", err)
	}

	log.Println("Test database connection closed successfully.")
}

func Setup(t *testing.T, app *fiber.App) *gorm.DB {
	// Set up the test database
	db := SetupTestDatabase()
	routes.SetupRoutes(app, db)
	DropData(db)
	CreateTables(db)
	return db
}

func ParseResponseBody(t *testing.T, b io.ReadCloser) interface{} {
	body, _ := io.ReadAll(b)
	// Parse the response body as JSON
	responseBody := make(map[string]interface{})
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		t.Errorf("error parsing response body as JSON: %s", err)
	}
	return responseBody
}

func ProcessTestBody(t *testing.T, app *fiber.App, url string, method string, body interface{}, access ...string) *http.Response {
	// Marshal the test data to JSON
	requestBytes, err := json.Marshal(body)
	requestBody := bytes.NewReader(requestBytes)
	assert.Nil(t, err)
	req := httptest.NewRequest(method, url, requestBody)
	req.Header.Set("Content-Type", "application/json")
	if access != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access[0]))
	}
	res, err := app.Test(req)
	if err != nil {
		log.Println(err)
	}
	return res
}
