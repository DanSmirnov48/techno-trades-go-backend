package controllers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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

// LoginUser handles the login logic
func LoginUser(c *fiber.Ctx) error {
	// 1) Parse and validate the login input.
	input, err := validate.ParseLoginInput(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User

	// 2) Check if user exists and retrieve their password
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect email or password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking user credentials"})
	}

	// 3) Compare the provided password with the stored password using the ComparePassword method
	if !user.ComparePassword(input.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect email or password"})
	}

	// 4) Create a JWT token
	accessToken, err := CreateToken(user.ID.String(), "24h")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating token"})
	}

	// 5) Set the token in a cookie
	isSecure := false
	if proto, ok := c.GetReqHeaders()["X-Forwarded-Proto"]; ok {
		for _, p := range proto {
			if strings.ToLower(p) == "https" {
				isSecure = true
				break
			}
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "strict",
	})

	// 6) Remove password from output
	user.Password = "" // Don't send the password back

	// 7) Send response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"token":  accessToken,
		"data": fiber.Map{
			"user": user,
		},
	})
}

// LogoutUser handles the user logout
func LogoutUser(c *fiber.Ctx) error {
	// Clear the access token cookie
	c.ClearCookie("accessToken")

	// Send response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// DecodeJWT verifies the JWT token, extracts the user ID, and retrieves the user from the database.
func DecodeJWT(c *fiber.Ctx) error {
	// Get the JWT from the cookies
	tokenString := c.Cookies("accessToken")

	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing or invalid JWT",
		})
	}

	// Parse the JWT token
	secret := os.Getenv("JWT_SECRET")
	token, err := ParseJWT(tokenString, secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate the token claims
	userID, err := ValidateJWTClaims(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Call the GetUserByID function
	user, err := GetUserByID(database.DB, userID)
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Attach the user object to the context
	c.Locals("user", user)

	// Return the user object in the response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"user": user,
		},
	})
}

// Middleware to attach the user object to the context
func Protect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the Authorization token from Header
		tokenString, err := GetAuthorizationHeader(c)
		if err != nil {
			fiberErr := err.(*fiber.Error)
			return c.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
		}

		// Parse the JWT token
		secret := os.Getenv("JWT_SECRET")
		token, err := ParseJWT(tokenString, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired token",
			})
		}

		// Validate the JWT claims
		userID, err := ValidateJWTClaims(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		user, err := GetUserByID(database.DB, userID)

		// Attach the user object to the context
		c.Locals("user", user)

		// Proceed to the next handler
		return c.Next()
	}
}

// Example handler to access the user object
func ProtectedEndpoint(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)

	if !ok {
		// If the user does not exist in Locals, set it to nil
		user = nil
	}

	// Log the request body
	fmt.Print(user)

	return c.JSON(fiber.Map{
		"user": user,
	})
}
