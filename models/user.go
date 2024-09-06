package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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
	Key  uuid.UUID `gorm:"type:uuid;"`
	Name string    `gorm:"size:255"`
	URL  string    `gorm:"size:255"`
}

type User struct {
	ID                           uuid.UUID `gorm:"type:uuid;primaryKey"`
	FirstName                    string    `gorm:"size:100;not null"`
	LastName                     string    `gorm:"size:100;not null"`
	Email                        string    `gorm:"size:100;unique;not null"`
	Role                         Role      `gorm:"size:50;not null"`
	Photo                        *Photo    `gorm:"embedded;embeddedPrefix:photo_"`
	Password                     string    `gorm:"size:255;not null"`
	Active                       bool      `gorm:"default:true"`
	Verified                     bool      `gorm:"default:false"`
	VerificationCode             int64
	EmailUpdateVerificationToken string `gorm:"size:255"`
	PasswordResetToken           string `gorm:"size:255"`
	PasswordResetTokenExpires    time.Time
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
	DeletedAt                    gorm.DeletedAt `gorm:"index"`
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

// BeforeUpdate is a GORM hook that runs before saving the User model
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	fmt.Println("Running BeforeUpdate function.")

	if tx.Statement.Changed("Password") {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		tx.Statement.SetColumn("Password", string(hashedPassword))
	}
	return nil
}

// AfterUpdate is a GORM hook that runs before saving the User model
func (u *User) AfterUpdate(tx *gorm.DB) (err error) {

	fmt.Println("Running AfterUpdate function.")

	return nil
}

// CreatePasswordResetVerificationCode generates a password reset code, sets the token and expiration time
func (u *User) CreatePasswordResetVerificationToken() (string, error) {
	// Generate a 4-byte random token
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := strings.ToUpper(hex.EncodeToString(bytes))

	// Set the token as the password reset token
	u.PasswordResetToken = token

	// Set the expiration time to 10 minutes from now
	u.PasswordResetTokenExpires = time.Now().Add(10 * time.Minute)

	return token, nil
}

// CreatePasswordResetVerificationCode generates a password reset code, sets the token and expiration time
func (u *User) CreateEmailUpdateVerificationToken() (string, error) {
	// Generate a 4-byte random token
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := strings.ToUpper(hex.EncodeToString(bytes))

	// Set the token as the password reset token
	u.EmailUpdateVerificationToken = token

	return token, nil
}
