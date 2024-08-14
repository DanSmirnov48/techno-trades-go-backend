package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
)

type Photo struct {
	Key  string `gorm:"size:255"`
	Name string `gorm:"size:255"`
	URL  string `gorm:"size:255"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	FirstName string    `gorm:"size:100;not null"`
	LastName  string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:100;unique;not null"`
	Role      Role      `gorm:"size:50;not null"`
	Photo     *Photo    `gorm:"embedded;embeddedPrefix:photo_"`
	Password  string    `gorm:"size:255;not null"`
	Active    bool      `gorm:"default:true"`
	Verified  bool      `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate is a GORM hook that runs before a User is created
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.Role = UserRole

	fmt.Println("Running BeforeCreate function.")

	return
}

// AfterCreate is a GORM hook that runs after a User is created
func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	// Log the details of the created user
	fmt.Println("Running AfterCreate function.")

	return
}

func (u *User) BeforeDelete(tx *gorm.DB) (err error) {
	// Log the details of the user to be deleted
	fmt.Println("Running BeforeDelete function.")

	return
}

func (u *User) AfterDelete(tx *gorm.DB) (err error) {
	// Log the details of the deleted user
	fmt.Println("Running AfterDelete function.")

	return
}
