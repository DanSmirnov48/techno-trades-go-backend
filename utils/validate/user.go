package validate

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// LoginInput holds the input data for logging in a user.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ParseLoginInput parses and validates the login input from the request body.
func ParseLoginInput(c *fiber.Ctx) (*LoginInput, error) {
	// Unmarshal into a map first
	var inputMap map[string]interface{}
	if err := c.BodyParser(&inputMap); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}

	// Type assertions and validations
	email, ok := inputMap["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email is required and must be a non-empty string")
	}

	password, ok := inputMap["password"].(string)
	if !ok || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("password is required and must be a non-empty string")
	}

	// Create and return a validated LoginInput
	input := &LoginInput{
		Email:    strings.TrimSpace(email),
		Password: strings.TrimSpace(password),
	}

	return input, nil
}
