package validate

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//---------------------USER LOGIN VALIDATION----------------------------------

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

//---------------------CREATE USER VALIDATION----------------------------------

// CreateUserInput holds the input data for creating a new user.
type CreateUserInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// ParseCreateUserInput parses and validates the input for creating a new user.
func ParseCreateUserInput(c *fiber.Ctx) (*CreateUserInput, error) {
	// Unmarshal into a map first
	var inputMap map[string]interface{}
	if err := c.BodyParser(&inputMap); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}

	// Type assertions and validations
	firstName, ok := inputMap["firstName"].(string)
	if !ok || strings.TrimSpace(firstName) == "" {
		return nil, fmt.Errorf("first name is required and must be a non-empty string")
	}

	lastName, ok := inputMap["lastName"].(string)
	if !ok || strings.TrimSpace(lastName) == "" {
		return nil, fmt.Errorf("last name is required and must be a non-empty string")
	}

	email, ok := inputMap["email"].(string)
	if !ok || strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email is required and must be a non-empty string")
	}

	password, ok := inputMap["password"].(string)
	if !ok || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("password is required and must be a non-empty string")
	}

	if !isEmailValid(email) {
		return nil, fmt.Errorf("email is not a valid email address")
	}

	// Create and return a validated CreateUserInput
	input := &CreateUserInput{
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
		Email:     strings.TrimSpace(email),
		Password:  strings.TrimSpace(password),
	}

	return input, nil
}

func isEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
