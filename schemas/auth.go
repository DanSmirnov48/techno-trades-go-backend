package schemas

import "github.com/DanSmirnov48/techno-trades-go-backend/models"

// REQUEST BODY SCHEMAS
type LoginSchema struct {
	Email    string `json:"email" validate:"required,email" example:"johndoe@email.com"`
	Password string `json:"password" validate:"required" example:"password"`
}

// RESPONSE BODY SCHEMAS
type TokensResponseSchema struct {
	User   *models.User `json:"user"`
	Access string       `json:"access"`
}

type LoginResponseSchema struct {
	ResponseSchema
	Data TokensResponseSchema `json:"data"`
}
