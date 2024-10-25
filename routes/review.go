package routes

import (
	"encoding/json"
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
)

var (
	reviewManager = managers.ReviewManager{}
)

func (endpoint Endpoint) CreateNewReview(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	reqData := schemas.CreateReview{}

	productId, err := utils.ParseUUID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(err)
	}

	if errCode, errData := ValidateRequest(c, &reqData); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	review, errCode, errData := reviewManager.Create(db, reqData, &user.ID, productId)
	if errCode != nil {
		return c.Status(*errCode).JSON(errData)
	}

	data, _ := json.MarshalIndent(&review, "", "  ")
	fmt.Println(string(data))

	response := schemas.ReviewResponseSchema{
		ResponseSchema: SuccessResponse("Review created successfully"),
		Data:           schemas.NewReviewResponseSchema{Review: review},
	}
	return c.Status(201).JSON(response)
}
