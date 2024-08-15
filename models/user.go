package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

// ComparePassword compares the hashed password with a plain password
func (u *User) ComparePassword(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	return err == nil
}

// BeforeCreate is a GORM hook that runs before a User is created
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	u.Role = UserRole

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

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
