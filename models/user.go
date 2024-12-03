package models

import (
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthType string

const (
	AuthTypePassword AuthType = "Password"
	AuthTypeGoogle   AuthType = "Google"
)

type AccountType string

const (
	AccountTypeBuyer AccountType = "Buyer"
	AccountTypeStaff AccountType = "Staff"
)

type User struct {
	ID              uuid.UUID      `json:"id,omitempty" gorm:"type:uuid;primarykey;not null;default:uuid_generate_v4()" example:"d10dde64-a242-4ed0-bd75-4c759644b3a6"`
	FirstName       string         `json:"first_name" gorm:"type: varchar(255);not null" example:"John"`
	LastName        string         `json:"last_name" gorm:"type: varchar(255);not null" example:"Doe"`
	Email           string         `json:"email" gorm:"not null;unique;" example:"johndoe@email.com"`
	Avatar          *string        `json:"avatar" gorm:"nullable"`
	Password        string         `json:"password" gorm:"not null"`
	IsEmailVerified bool           `json:"-" gorm:"default:false"`
	AuthType        AuthType       `json:"authType" gorm:"type:varchar(50);default:'Password'"`
	AccountType     AccountType    `json:"accountType" gorm:"type:varchar(50);default:'Buyer'"`
	Active          bool           `json:"-" gorm:"default:true"`
	Access          *string        `gorm:"type:varchar(1000);null;" json:"-"`
	Refresh         *string        `gorm:"type:varchar(1000);null;" json:"-"`
	Products        []Product      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	CreatedAt       time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	hashedPassword := utils.HashPassword(u.Password)
	u.Password = hashedPassword

	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("Password") {
		tx.Statement.SetColumn("Password", utils.HashPassword(u.Password))
	}
	return nil
}
