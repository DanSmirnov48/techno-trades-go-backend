package routes

import (
	"encoding/json"
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	productManager = managers.ProductManager{}
)

func (endpoint Endpoint) CreateNewProduct(c *fiber.Ctx) error {
	db := endpoint.DB
	createProductSchema := schemas.CreateProduct{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &createProductSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	user, ok := c.Locals("user").(*models.User)
	if !ok || user == nil || user.Role != models.AdminRole {
		return c.Status(401).JSON(utils.RequestErr(utils.ERR_UNAUTHORIZED_USER, "Unauthorized Access"))
	}

	newProduct := productManager.Create(db, createProductSchema, user.ID)
	if newProduct.ID == uuid.Nil {
		return c.Status(404).JSON(utils.RequestErr(utils.ERR_NETWORK_FAILURE, "Error creating product"))
	}

	data, _ := json.MarshalIndent(&newProduct, "", "  ")
	fmt.Println(string(data))

	response := schemas.ProductCreateResponseSchema{
		ResponseSchema: SuccessResponse("Product created successfully"),
		Data:           schemas.NewProductResponseSchema{Product: newProduct},
	}
	return c.Status(201).JSON(response)
}
