package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// CreateToken generates a JWT token for the given user ID
func CreateToken(userID string, expiresIn string) (string, error) {
	// Parse the expiration duration
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return "", err
	}

	// Define claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	}

	// Create token
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseJWT parses the JWT token from a string and returns the token object.
func ParseJWT(tokenString string, secret string) (*jwt.Token, error) {
	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}

// ValidateJWTClaims validates the JWT claims and checks for expiration.
func ValidateJWTClaims(token *jwt.Token) (string, error) {
	// Check if the token is valid
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	// Extract the user ID from the token
	userID, ok := claims["user_id"].(string) // Assuming "sub" is the field for user ID
	if !ok {
		return "", fmt.Errorf("invalid token: missing user ID")
	}

	// Check if the token is expired
	expiration, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(expiration) {
		return "", fmt.Errorf("token expired")
	}

	return userID, nil
}

// ExtractToken extracts the token from the Authorization header.
func GetAuthorizationHeader(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "No authorization header provided")
	}

	// Trim the "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Authorization header format must be Bearer {token}")
	}

	return tokenString, nil
}

// FilteredFields filters the fields that are allowed to be updated.
func FilteredFields(body map[string]interface{}, allowedFields ...string) map[string]interface{} {
	filtered := make(map[string]interface{})
	for _, field := range allowedFields {
		if value, exists := body[field]; exists {
			filtered[field] = value
		}
	}
	return filtered
}

// GenerateRandomToken generates a random token of specified length and case (upper or lower).
func GenerateRandomToken(length int, uppercase bool) (string, error) {
	// Generate a byte slice of half the length since each byte represents 2 hex characters.
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert the bytes to a hex string.
	token := hex.EncodeToString(bytes)

	// Convert the token to uppercase if specified.
	if uppercase {
		token = strings.ToUpper(token)
	} else {
		token = strings.ToLower(token)
	}

	return token, nil
}
