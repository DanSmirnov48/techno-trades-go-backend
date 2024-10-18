package models

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Otp struct {
	ID        uuid.UUID `json:"id,omitempty" gorm:"type:uuid;primarykey;not null;default:uuid_generate_v4()" example:"d10dde64-a242-4ed0-bd75-4c759644b3a6"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	UserId    uuid.UUID `json:"user_id" gorm:"unique"`
	User      User      `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE"`
	Code      uint32    `json:"code"`
}

func (otp *Otp) BeforeSave(tx *gorm.DB) (err error) {
	otp.Code = generateOtpCode()
	expirationDuration := config.GetConfig().EmailOtpExpireMins
	otp.ExpiresAt = time.Now().Add(time.Minute * time.Duration(expirationDuration))
	return
}

func (obj Otp) CheckExpiration() bool {
	return time.Now().After(obj.ExpiresAt)
}

func generateOtpCode() uint32 {
	max := int64(1000000)
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	return uint32(n.Int64())
}
