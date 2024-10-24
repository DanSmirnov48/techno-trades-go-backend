package schemas

import (
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/google/uuid"
)

// REQUEST BODY SCHEMAS
type CreateReview struct {
	Title     string     `json:"title" validate:"required,max=50" example:"Good quality product"`
	Comment   string     `json:"comment" validate:"required,max=500"`
	Rating    int        `json:"rating" validate:"required,min=0,max=5" example:"5"`
	ProductId *uuid.UUID `json:"product_id" validate:"omitempty" example:"d10dde64-a242-4ed0-bd75-4c759644b3a6"`
}

// RESPONSE BODY SCHEMAS
type NewReviewResponseSchema struct {
	Review *models.Review `json:"review"`
}

type ReviewResponseSchema struct {
	ResponseSchema
	Data NewReviewResponseSchema `json:"data"`
}
