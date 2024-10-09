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

type LoginResponseSchema struct {
	ResponseSchema
	Data TokensResponseSchema `json:"data"`
}
