package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/database"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils/validate"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// LogIn handles the login logic
func LogIn(c *fiber.Ctx) error {
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
	accessToken, err := utils.CreateToken(user.ID.String(), "24h")
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

// LogOut handles the user logout
func LogOut(c *fiber.Ctx) error {
	// Clear the access token cookie
	c.ClearCookie("accessToken")

	// Send response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

// DecodeJWT verifies the JWT token, extracts the user ID, and retrieves the user from the database.
func GetCurrentUser(c *fiber.Ctx) error {
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
	token, err := utils.ParseJWT(tokenString, secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate the token claims
	userID, err := utils.ValidateJWTClaims(token)
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

// DecodeJWT verifies the JWT token, extracts the user ID, and retrieves the user from the database.
func Validate(c *fiber.Ctx) error {
	// Get the JWT from the cookies
	accessToken := c.Cookies("accessToken")

	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing or invalid JWT",
		})
	}

	// Parse the JWT token
	secret := os.Getenv("JWT_SECRET")
	token, err := utils.ParseJWT(accessToken, secret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate the token claims
	userID, err := utils.ValidateJWTClaims(token)
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
		"token":  accessToken,
		"data": fiber.Map{
			"user": user,
		},
	})
}

// UpdateUserPassword allows authenticated users to update their password.
func UpdateUserPassword(c *fiber.Ctx) error {
	// UpdatePasswordInput holds the data for updating the user's password.
	type UpdatePasswordInput struct {
		CurrentPassword string `json:"currentPassword" validate:"required"`
		NewPassword     string `json:"newPassword" validate:"required,min=8"`
	}

	// Parse the request body into the UpdatePasswordInput struct
	var input UpdatePasswordInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get the authenticated user from the context
	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "User information is missing. You do not have permission to perform this action.",
		})
	}

	// Compare the provided password with the stored password using the ComparePassword method
	if !user.ComparePassword(input.CurrentPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Current password is incorrect",
		})
	}

	// Update the user's password (it will be hashed in the BeforeSave hook)
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"Password": input.NewPassword,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update password",
		})
	}

	// Return a success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Password updated successfully",
	})
}

func ForgotPassword(c *fiber.Ctx) error {
	type ForgotPassword struct {
		Email string `json:"email"`
	}

	var input ForgotPassword
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	var user models.User

	// Check if user exists and retrieve their password
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect email or password"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking user credentials"})
	}

	// Generate a password reset token
	token, err := user.CreatePasswordResetVerificationToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate password reset code",
		})
	}

	// Save the changes to the database
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to save password reset token",
		})
	}

	// TODO: EMAIL THE TOKEN TO THE USER!!

	// Return a success response with the code
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Password reset code has been sent to your email.",
		"token":   token,
	})
}

func VerifyPasswordResetToken(c *fiber.Ctx) error {
	// Input structure to capture the token from the request body
	type VerifyTokenInput struct {
		Token string `json:"token" validate:"required"`
	}

	var input VerifyTokenInput

	// Parse the request body into the VerifyTokenInput struct
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	var user models.User

	// Retrieve the user by PasswordResetToken and check if the token is not expired
	if err := database.DB.Where("password_reset_token = ? AND password_reset_token_expires > ?", input.Token, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired token",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error checking user credentials",
		})
	}

	// If the token is valid, allow the user to proceed with resetting their password
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Token is valid, you may proceed to reset your password",
	})
}

func ResetUserPassword(c *fiber.Ctx) error {
	// Input structure to capture the new password, email, and token
	type ResetPasswordInput struct {
		Email    string `json:"email" validate:"required,email"`
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	var input ResetPasswordInput

	// Parse the request body into the ResetPasswordInput struct
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	var user models.User

	// Retrieve the user by email and verify the token and expiration time
	if err := database.DB.Where("email = ? AND password_reset_token = ? AND password_reset_token_expires > ?", input.Email, input.Token, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired token",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error checking user credentials",
		})
	}

	// Update the user's password (it will be hashed in the BeforeSave hook)
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"Password":                  input.Password,
		"PasswordResetToken":        nil,
		"PasswordResetTokenExpires": nil,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update password",
		})
	}

	// Return a success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Password updated successfully",
		"user":    user,
	})
}
