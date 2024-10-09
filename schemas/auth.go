package schemas

import "github.com/DanSmirnov48/techno-trades-go-backend/models"

// REQUEST BODY SCHEMAS
type LoginSchema struct {
	Email    string `json:"email" validate:"required,email" example:"johndoe@email.com"`
	Password string `json:"password" validate:"required" example:"password"`
}

type RegisterUser struct {
	FirstName        string `json:"first_name" validate:"required,max=50" example:"John"`
	LastName         string `json:"last_name" validate:"required,max=50" example:"Doe"`
	Email            string `json:"email" validate:"required,min=5,email" example:"johndoe@email.com"`
	Password         string `json:"password" validate:"required,min=8,max=50" example:"strongpassword"`
	VerificationCode int64  `json:"verification_code"`
}

type EmailRequestSchema struct {
	Email string `json:"email" validate:"required,min=5,email" example:"johndoe@email.com"`
}

type VerifyAccountRequestSchema struct {
	Email            string `json:"email" validate:"required,min=5,email" example:"johndoe@example.com"`
	VerificationCode int64  `json:"verification_code" validate:"required" example:"123456"`
}

type RefreshTokenSchema struct {
	Refresh string `json:"refresh" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InNpbXBsZWlkIiwiZXhwIjoxMjU3ODk0MzAwfQ.Ys_jP70xdxch32hFECfJQuvpvU5_IiTIN2pJJv68EqQ"`
}

// RESPONSE BODY SCHEMAS
type RegisterResponseSchema struct {
	ResponseSchema
	Data EmailRequestSchema `json:"data"`
}

type TokensResponseSchema struct {
	User    *models.User `json:"user"`
	Access  string       `json:"access"`
	Refresh string       `json:"refresh"`
}

type MagicLinkResponseSchema struct {
	Link string `json:"link"`
}

type LoginResponseSchema struct {
	ResponseSchema
	Data TokensResponseSchema `json:"data"`
}

type MagicLinkLoginResponseSchema struct {
	ResponseSchema
	Data MagicLinkResponseSchema `json:"data"`
}
