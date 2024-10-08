// config/config.go

package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

// Configuration holds the application configuration loaded from environment variables
type Configuration struct {
	EmailOTPExpireSeconds     int64
	AccessTokenExpireMinutes  int
	RefreshTokenExpireMinutes int
	SecretKey                 string
	FrontendURL               string
	PostgresUser              string
	PostgresPassword          string
	PostgresServer            string
	PostgresPort              string
	PostgresDB                string
	MailSenderEmail           string
	MailSenderPassword        string
	MailSenderHost            string
	MailSenderPort            int
	CORSAllowedOrigins        string
}

var config *Configuration

func init() {
	// Load environment variables from the .env file (if it exists) into the environment
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Unable to identify current directory (needed to load .env)", os.Stderr)
		os.Exit(1)
	}
	basepath := filepath.Dir(file)
	err := godotenv.Load(filepath.Join(basepath, "../.env"))

	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Convert string-based numeric variables to their respective types
	emailOTPExpireSeconds, _ := strconv.ParseInt(os.Getenv("EMAIL_OTP_EXPIRE_SECONDS"), 10, 64)
	mailSenderPort, _ := strconv.Atoi(os.Getenv("MAIL_SENDER_PORT"))
	accessTokenExpireMinutes, _ := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES"))
	refreshTokenExpireMinutes, _ := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRE_MINUTES"))

	config = &Configuration{
		EmailOTPExpireSeconds:     emailOTPExpireSeconds,
		AccessTokenExpireMinutes:  accessTokenExpireMinutes,
		RefreshTokenExpireMinutes: refreshTokenExpireMinutes,
		SecretKey:                 os.Getenv("SECRET_KEY"),
		FrontendURL:               os.Getenv("CLIENT_URL"),
		PostgresUser:              os.Getenv("POSTGRES_USER"),
		PostgresPassword:          os.Getenv("POSTGRES_PASSWORD"),
		PostgresServer:            os.Getenv("POSTGRES_SERVER"),
		PostgresPort:              os.Getenv("POSTGRES_PORT"),
		PostgresDB:                os.Getenv("POSTGRES_DB"),
		MailSenderEmail:           os.Getenv("MAIL_SENDER_EMAIL"),
		MailSenderPassword:        os.Getenv("MAIL_SENDER_PASSWORD"),
		MailSenderHost:            os.Getenv("MAIL_SENDER_HOST"),
		MailSenderPort:            mailSenderPort,
		CORSAllowedOrigins:        os.Getenv("CORS_ALLOWED_ORIGINS"),
	}
}

// GetConfig returns the application configuration
func GetConfig() *Configuration {
	return config
}
