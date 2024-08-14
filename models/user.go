package models

import (
	"time"

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
	ID        uint   `gorm:"primaryKey"`
	FirstName string `gorm:"size:100;not null"`
	LastName  string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;not null;unique"`
	Role      Role   `gorm:"size:50;not null"`
	Photo     *Photo `gorm:"embedded;embeddedPrefix:photo_"`
	Password  string `gorm:"size:255;not null"`
	Active    bool   `gorm:"default:true"`
	Verified  bool   `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
