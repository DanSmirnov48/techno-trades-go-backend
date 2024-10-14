package managers

import (
	"fmt"
	"log"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ----------------------------------
// USER MANAGEMENT
// --------------------------------
type UserManager struct{}

func (obj UserManager) GetById(db *gorm.DB, id uuid.UUID) (*models.User, *fiber.Error) {
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}
	return &user, nil
}

func (obj UserManager) GetAll(db *gorm.DB) ([]models.User, *fiber.Error) {
	var users []models.User

	if err := db.Find(&users).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	return users, nil
}

func (obj UserManager) GetByEmail(db *gorm.DB, email string) (*models.User, *fiber.Error) {
	var user models.User

	if err := db.First(&user, "email = ?", email).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	return &user, nil
}

func (obj UserManager) GetByMagicLoginToken(db *gorm.DB, token string) (*models.User, *fiber.Error) {
	var user models.User

	if err := db.Where("magic_log_in_token = ? AND magic_log_in_token_expires > ?", token, time.Now()).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "Refresh token is invalid or expired")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	return &user, nil
}

func (obj UserManager) Create(db *gorm.DB, userSchema schemas.RegisterUser, isEmailVerified bool, isAdmin bool) (*models.User, *fiber.Error) {

	role := models.UserRole
	if isAdmin {
		role = models.AdminRole
	}

	newUser := models.User{
		FirstName:        userSchema.FirstName,
		LastName:         userSchema.LastName,
		Email:            userSchema.Email,
		Password:         userSchema.Password,
		VerificationCode: utils.GetRandomInt(6),
		Verified:         isEmailVerified,
		Role:             role,
	}

	if err := db.Create(&newUser).Error; err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Could not create user")
	}

	return &newUser, nil
}

func (obj UserManager) GetOrCreate(db *gorm.DB, userData schemas.RegisterUser, isEmailVerified bool, isAdmin bool) *models.User {
	user, _ := obj.GetByEmail(db, userData.Email)
	if user == nil {
		user, _ = obj.Create(db, userData, isEmailVerified, isAdmin)
	}
	return user
}

func (obj UserManager) SetAccountVerified(db *gorm.DB, user *models.User) *fiber.Error {
	if err := db.Model(user).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"verified":          true,
			"verification_code": nil,
		}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user information")
	}
	return nil
}

func (obj UserManager) ClearMagicLogin(db *gorm.DB, user *models.User) {
	db.Model(&user).Updates(map[string]interface{}{
		"MagicLogInToken":        nil,
		"MagicLogInTokenExpires": nil,
	})
}

func (obj UserManager) DropData(db *gorm.DB) error {
	// Use the GORM Migrator to drop the User table
	if err := db.Migrator().DropTable(&models.User{}); err != nil {
		return fmt.Errorf("failed to drop user table: %w", err)
	}
	log.Println("User table dropped successfully.")
	return nil
}
