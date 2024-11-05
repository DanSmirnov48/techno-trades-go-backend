package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"gorm.io/gorm"
)

// Initialize Stripe
func init() {
	stripe.Key = config.GetConfig().StripeSecretKey
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateCheckOutSchema struct {
	Orders []OrderItem `json:"orders"`
}

func (endpoint Endpoint) CreateCheckoutSession(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	data := CreateCheckOutSchema{}
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	// Validate that orders array is not empty
	if len(data.Orders) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Orders cannot be empty"})
	}

	// Create a Stripe customer with metadata
	params := &stripe.CustomerParams{
		Name:  stripe.String(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		Email: stripe.String(user.Email),
		Metadata: map[string]string{
			"userId": user.ID.String(),
			"cart":   serializeOrders(data.Orders),
		},
	}

	customer, err := customer.New(params)
	if err != nil {
		log.Printf("Stripe customer creation error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create customer"})
	}

	// Calculate total amount for verification
	var totalAmount int64 = 0

	// Prepare line items for each product in the order
	var lineItems []*stripe.CheckoutSessionLineItemParams
	for _, item := range data.Orders {
		// Validate quantity
		if item.Quantity <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Quantity must be greater than 0"})
		}

		// Fetch product details from your database
		var dbProduct models.Product
		if err := db.First(&dbProduct, "id = ?", item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("Product not found for ID: %s", item.ProductID)})
			}
			log.Printf("Database error when fetching product: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}

		// Check stock availability
		if item.Quantity > dbProduct.CountInStock {
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Insufficient stock for product: %s. Available: %d, Requested: %d",
					dbProduct.Name, dbProduct.CountInStock, item.Quantity),
			})
		}

		// Add to total amount for verification
		totalAmount += int64(dbProduct.Price*100) * int64(item.Quantity)

		// Define line item with product data and quantity
		lineItem := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("gbp"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(dbProduct.Name),
					Metadata: map[string]string{
						"id":             dbProduct.ID.String(),
						"original_price": fmt.Sprintf("%d", dbProduct.Price),
					},
				},
				UnitAmount: stripe.Int64(int64(dbProduct.Price * 100)),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		}
		lineItems = append(lineItems, lineItem)
	}

	// Create the checkout session
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		Customer:           stripe.String(customer.ID),
		SuccessURL:         stripe.String(fmt.Sprintf("%s/checkout/success?session_id={CHECKOUT_SESSION_ID}", config.GetConfig().FrontendURL)),
		CancelURL:          stripe.String(fmt.Sprintf("%s/checkout/cancel", config.GetConfig().FrontendURL)),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"userId":      user.ID.String(),
				"totalAmount": fmt.Sprintf("%d", totalAmount),
				"orderCount":  fmt.Sprintf("%d", len(data.Orders)),
			},
		},
		Metadata: map[string]string{
			"userId":     user.ID.String(),
			"totalItems": fmt.Sprintf("%d", len(data.Orders)),
		},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		log.Printf("Stripe session creation error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create checkout session"})
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"url":     session.URL,
	})
}

func serializeOrders(orders []OrderItem) string {
	data, _ := json.Marshal(orders)
	return string(data)
}
