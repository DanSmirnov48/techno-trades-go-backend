package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Image struct {
	Key  uuid.UUID `gorm:"type:uuid;"`
	Name string    `gorm:"size:255"`
	URL  string    `gorm:"size:255"`
}

type Product struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug            string    `gorm:"size:255;not null"`
	Name            string    `gorm:"size:255;not null"`
	Image           []Image   `gorm:"type:json"`
	Brand           string    `gorm:"size:255;not null"`
	Category        string    `gorm:"size:255;not null"`
	Description     string    `gorm:"type:text"`
	Rating          float64   `gorm:"default:0;not null"`
	Price           float64   `gorm:"not null"`
	CountInStock    int       `gorm:"not null"`
	IsDiscounted    bool      `gorm:"default:false;not null"`
	DiscountedPrice *float64  `gorm:"default:null"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`                             // Foreign key to reference the user
	User            User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"` // Establishes relationship with User
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}
