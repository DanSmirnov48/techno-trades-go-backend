package routes

import (
	"encoding/json"
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

var (
	productManager = managers.ProductManager{}
)

func (endpoint Endpoint) CreateNewProduct(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	createProductSchema := schemas.CreateProduct{}

	// Validate request
	if errCode, errData := ValidateRequest(c, &createProductSchema); errData != nil {
		return c.Status(*errCode).JSON(errData)
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

func (endpoint Endpoint) GetAllProducts(c *fiber.Ctx) error {
	db := endpoint.DB

	products, errCode, errData := productManager.GetAll(db)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	response := schemas.FindManyProductsResponseSchem{Products: products, Length: len(products)}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) FindProductById(c *fiber.Ctx) error {
	db := endpoint.DB
	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	product, errCode, errData := productManager.GetById(db, *productId)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	response := schemas.FindSingleProductResponseSchem{Product: product}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) FindProductBySlug(c *fiber.Ctx) error {
	db := endpoint.DB
	slug := c.Params("slug")

	product, errCode, errData := productManager.GetBySlug(db, slug)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	response := schemas.FindSingleProductResponseSchem{Product: product}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) SetProductDiscount(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.UpdateDiscount{}

	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	updatedProduct, errCode, errData := productManager.UpdateDiscount(db, *productId, reqData)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	response := schemas.ProductCreateResponseSchema{
		ResponseSchema: SuccessResponse("Discount Updated successfully"),
		Data:           schemas.NewProductResponseSchema{Product: updatedProduct},
	}
	return c.Status(201).JSON(response)
}

func (endpoint Endpoint) UpdateProductStock(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.UpdateStockSchema{}

	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	updatedProduct, errCode, errData := productManager.UpdateStock(db, *productId, reqData.StockChange)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	response := schemas.ProductCreateResponseSchema{
		ResponseSchema: SuccessResponse("Stock updated successfully"),
		Data:           schemas.NewProductResponseSchema{Product: updatedProduct},
	}
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) UpdateProductDetails(c *fiber.Ctx) error {
	db := endpoint.DB
	reqData := schemas.UpdateProduct{}

	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	product, errCode, errData := productManager.GetById(db, *productId)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	if reqData.Name != product.Name {
		product.Slug = slug.Make(reqData.Name)
	}
	db.Model(&product).Updates(reqData)

	response := schemas.ProductCreateResponseSchema{
		ResponseSchema: SuccessResponse("Product updated successfully"),
		Data:           schemas.NewProductResponseSchema{Product: product},
	}
	return c.Status(200).JSON(response)
}

func (endpoint Endpoint) DeleteProduct(c *fiber.Ctx) error {
	db := endpoint.DB

	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	product, errCode, errData := productManager.GetById(db, *productId)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	db.Delete(&product)

	return c.Status(200).JSON(SuccessResponse("Product deleted successfully"))
}
