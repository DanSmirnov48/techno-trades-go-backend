package schemas

import "github.com/DanSmirnov48/techno-trades-go-backend/models"

// REQUEST BODY SCHEMAS
type CreateProduct struct {
	Name            string  `json:"name" validate:"required,max=50" example:"Sony PlayStation 5"`
	Brand           string  `json:"brand" validate:"required,max=50" example:"Sony"`
	Category        string  `json:"category" validate:"required" example:"consoles"`
	Description     string  `json:"description" validate:"required,max=500" example:"some item description"`
	Price           float64 `json:"price" validate:"required,gt=0" example:"399.99"`
	CountInStock    int     `json:"stock" validate:"required,min=0" example:"100"`
	IsDiscounted    bool    `json:"is_discounted"`
	DiscountedPrice float64 `json:"discounted_price" validate:"discounted_price_valid" example:"299.99"`
}

// RESPONSE BODY SCHEMAS
type NewProductResponseSchema struct {
	Product *models.Product `json:"product"`
}

type ProductCreateResponseSchema struct {
	ResponseSchema
	Data NewProductResponseSchema `json:"data"`
}

type FindManyProductsResponseSchem struct {
	Products []*models.Product `json:"products"`
	Length   int               `json:"length"`
}

type FindSingleProductResponseSchem struct {
	Product *models.Product `json:"product"`
}
