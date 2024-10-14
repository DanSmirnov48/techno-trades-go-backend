package tests

import (
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"gorm.io/gorm"
)

var (
	userManager = managers.UserManager{}
)

func CreateTestUser(db *gorm.DB) *models.User {
	rndEmail := fmt.Sprintf("%s@example.com", utils.GetRandomString(10))

	userData := schemas.RegisterUser{
		FirstName: "Test",
		LastName:  "User",
		Email:     rndEmail,
		Password:  "testpassword",
	}
	user := userManager.GetOrCreate(db, userData, false, false)
	return user
}

func CreateVerifiedTestUser(db *gorm.DB) *models.User {

	rndEmail := fmt.Sprintf("testverifieduser%s@example.com", utils.GetRandomString(5))

	userData := schemas.RegisterUser{
		FirstName: "Test",
		LastName:  "Verified",
		Email:     rndEmail,
		Password:  "testpassword",
	}
	user := userManager.GetOrCreate(db, userData, true, false)
	return user
}
