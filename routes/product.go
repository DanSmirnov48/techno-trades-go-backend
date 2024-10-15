package routes

import (
	"encoding/json"
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

func (endpoint Endpoint) CreateNewProduct(c *fiber.Ctx) error {
	db := endpoint.DB
	createProductSchema := schemas.CreateProduct{}

	// Validate request
	if errCode, errData := DecodeJSONBody(c, &createProductSchema); errData != nil {
		return c.Status(errCode).JSON(errData)
	}
	if err := validator.Validate(createProductSchema); err != nil {
		return c.Status(422).JSON(err)
	}

	uid, _ := utils.ParseUUID("39142de6-c6ec-4239-9dff-c192eea53f90")

	user, _ := userManager.GetById(db, *uid)
	if user == nil {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "Invalid Credentials"))
	}

	product := models.Product{
		ID:              uuid.New(),
		Name:            createProductSchema.Name,
		Slug:            slug.Make(createProductSchema.Name),
		Brand:           createProductSchema.Brand,
		Category:        createProductSchema.Category,
		Description:     createProductSchema.Description,
		Rating:          0,
		Price:           createProductSchema.Price,
		CountInStock:    createProductSchema.CountInStock,
		IsDiscounted:    createProductSchema.IsDiscounted,
		DiscountedPrice: createProductSchema.DiscountedPrice,
		UserID:          user.ID,
	}

	if err := db.Create(&product).Error; err != nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, err.Error()))
	}

	data, _ := json.MarshalIndent(&product, "", "  ")
	fmt.Println(string(data))

	return c.Status(201).JSON("OK")
}
