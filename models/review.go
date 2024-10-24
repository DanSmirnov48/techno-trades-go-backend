package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID `json:"id,omitempty" gorm:"type:uuid;primarykey;not null;default:uuid_generate_v4()" example:"d10dde64-a242-4ed0-bd75-4c759644b3a6"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
	Title     string    `json:"title" gorm:"type: varchar(255);not null" example:"Good quality product"`
	Comment   string    `json:"comment" gorm:"type:varchar(1000);not null"`
	Rating    int       `json:"rating" validate:"required,min=0,max=5" example:"5"`
	UserId    uuid.UUID `json:"user_id" gorm:"unique"`
	User      User      `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE"`
	ProductId uuid.UUID `json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductId;constraint:OnDelete:CASCADE"`
}
