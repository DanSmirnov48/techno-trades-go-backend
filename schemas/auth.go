package schemas

type LoginSchema struct {
	Email    string `json:"email" validate:"required,email" example:"johndoe@email.com"`
	Password string `json:"password" validate:"required" example:"password"`
}
