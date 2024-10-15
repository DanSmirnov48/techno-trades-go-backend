package schemas

import "github.com/DanSmirnov48/techno-trades-go-backend/models"

// REQUEST BODY SCHEMAS
type PasswordResetOtpRequestSchema struct {
	Otp string `json:"otp" validate:"required" example:"ABC123"`
}

type UserPasswordResetRequestSchema struct {
	Email       string `json:"email" validate:"required,min=5,email" example:"johndoe@email.com"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=50" example:"strongpassword"`
	Otp         string `json:"otp" validate:"required" example:"ABC123"`
}

type UpdateUserPasswordRequestSchema struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=50" example:"strongpassword"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=50" example:"strongpassword"`
}

type UpdateUserEmailRequestSchema struct {
	Otp      string `json:"otp" validate:"required" example:"ABC123"`
	NewEmail string `json:"new_email" validate:"required,min=5,email" example:"johndoe@example.com"`
}

type UpdateUserRequestSchema struct {
	FirstName string `json:"first_name" validate:"required,max=50" example:"John"`
	LastName  string `json:"last_name" validate:"required,max=50" example:"Doe"`
}

// RESPONSE BODY SCHEMAS
type PasswordResetOtpResponseSchema struct {
	Email string `json:"email" validate:"required,min=5,email" example:"johndoe@example.com"`
	Otp   string `json:"otp" validate:"required" example:"ABC123"`
}

type SendPasswordResetOtpResponseSchema struct {
	ResponseSchema
	Data PasswordResetOtpResponseSchema `json:"data"`
}

type UserResponseSchem struct {
	Users *models.User `json:"users"`
}

type FindUserByIdResponseSchem struct {
	ResponseSchema
	Data UserResponseSchem `json:"data"`
}

type UsersResponseSchem struct {
	Users  []*models.User `json:"users"`
	Length int            `json:"length"`
}

type FindAllUsersResponseSchem struct {
	ResponseSchema
	Data UsersResponseSchem `json:"data"`
}
