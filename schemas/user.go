package schemas

import "github.com/DanSmirnov48/techno-trades-go-backend/models"

// REQUEST BODY SCHEMAS
type UpdateUserPasswordRequestSchema struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=50" example:"strongpassword"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=50" example:"strongpassword"`
}

type UpdateUserEmailRequestSchema struct {
	Otp      uint32 `json:"otp" validate:"required" example:"112233"`
	NewEmail string `json:"new_email" validate:"required,min=5,email" example:"johndoe@example.com"`
}

type UpdateUserRequestSchema struct {
	FirstName string `json:"first_name" validate:"max=50" example:"John"`
	LastName  string `json:"last_name" validate:"max=50" example:"Doe"`
}

// RESPONSE BODY SCHEMAS
type PasswordResetOtpResponseSchema struct {
	Email string `json:"email" validate:"required,min=5,email" example:"johndoe@example.com"`
	Otp   uint32 `json:"otp" validate:"required" example:"112233"`
}

type SendPasswordResetOtpResponseSchema struct {
	ResponseSchema
	Data PasswordResetOtpResponseSchema `json:"data"`
}

type UserResponseSchem struct {
	Users *models.User `json:"users"`
}

type SingleUserResponseSchem struct {
	ResponseSchema
	Data UserResponseSchem `json:"data"`
}

type UsersResponseSchem struct {
	Users  []*models.User `json:"users"`
	Length int            `json:"length"`
}

type ManyUsersResponseSchem struct {
	ResponseSchema
	Data UsersResponseSchem `json:"data"`
}
