package managers

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

func (obj UserManager) GetAll(db *gorm.DB) ([]*models.User, *fiber.Error) {
	var users []*models.User

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

func (obj UserManager) Create(db *gorm.DB, data schemas.RegisterUser, isEmailVerified bool, isAdmin bool) *models.User {
	newUser := utils.ConvertStructData(data, models.User{}).(*models.User)
	newUser.IsEmailVerified = isEmailVerified
	newUser.ID = uuid.New()
	if isAdmin {
		newUser.Role = models.AdminRole
	}
	db.Create(&newUser)
	return newUser
}
