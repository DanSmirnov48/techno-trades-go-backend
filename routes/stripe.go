package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/managers"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/webhook"
	"gorm.io/gorm"
)

// Initialize Stripe
func init() {
	stripe.Key = config.GetConfig().StripeSecretKey
	productManager = managers.ProductManager{}
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

func (endpoint Endpoint) CreatePaymentIntent(c *fiber.Ctx) error {
	db := endpoint.DB
	user := RequestUser(c)
	data := CreateCheckOutSchema{}
	if errCode, errData := ValidateRequest(c, &data); errData != nil {
		return c.Status(*errCode).JSON(errData)
	}

	customer, err := customer.New(&stripe.CustomerParams{
		Name:  stripe.String(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		Email: stripe.String(user.Email),
		Metadata: map[string]string{
			"userId": user.ID.String(),
			"cart":   serializeOrders(data.Orders),
		},
	})

	if err != nil {
		log.Printf("Stripe customer creation error: %v", err)
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to create customer"))
	}

	// Calculate total amount
	var totalAmount int64 = 0
	for _, item := range data.Orders {

		productId, err := utils.ParseUUID(item.ProductID)
		if err != nil {
			return c.Status(400).JSON(err)
		}

		product, errCode, errData := productManager.GetById(db, *productId)
		if errCode != nil {
			return c.Status(*errCode).JSON(errData)
		}

		if item.Quantity > product.CountInStock {
			return c.Status(404).JSON(utils.RequestErr(utils.ERR_NON_EXISTENT, "Insufficient product stock"))
		}

		// Add to total amount
		totalAmount += int64(product.Price*100) * int64(item.Quantity)
	}

	// Create a PaymentIntent with amount and currency
	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:             stripe.Int64(totalAmount),
		Customer:           stripe.String(customer.ID),
		Currency:           stripe.String(string(stripe.CurrencyGBP)),
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "paypal"}),
		Metadata: map[string]string{
			"userId":      user.ID.String(),
			"cart":        serializeOrders(data.Orders),
			"totalAmount": fmt.Sprintf("%d", totalAmount),
		},
	})
	if err != nil {
		log.Printf("Stripe session creation error: %v", err)
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_SERVER_ERROR, "Failed to create checkout session"))
	}

	return c.Status(200).JSON(fiber.Map{
		"success":      true,
		"clientSecret": pi.ClientSecret,
	})
}

func serializeOrders(orders []OrderItem) string {
	data, _ := json.Marshal(orders)
	return string(data)
}

func (endpoint Endpoint) HandleStripeWebhook(c *fiber.Ctx) error {
	// Read the request body
	webhookSecret := config.GetConfig().StripeTestKey
	stripeSignature := c.Get("Stripe-Signature")
	body := c.Body()

	// Verify webhook signature
	event, err := webhook.ConstructEvent(body, stripeSignature, webhookSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Webhook signature verification failed"})
	}

	// Handle the checkout.session.completed event
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": "Error parsing webhook data"})
		}

		// Print the full session object for inspection
		prettyJSON, _ := json.MarshalIndent(session, "", "    ")
		log.Printf("Full Checkout Session:\n%s\n", string(prettyJSON))

		// Print specific important fields
		log.Printf("\n=== Checkout Session Summary ===")
		log.Printf("Session ID: %s", session.ID)
		log.Printf("Customer ID: %s", session.Customer.ID)
		log.Printf("Customer Email: %s", session.CustomerEmail)
		log.Printf("Payment Status: %s", session.PaymentStatus)
		log.Printf("Amount Total: %d", session.AmountTotal)
		log.Printf("Currency: %s", session.Currency)

		// Print metadata
		log.Printf("\n=== Metadata ===")
		for key, value := range session.Metadata {
			log.Printf("%s: %s", key, value)
		}

		// Print line items
		log.Printf("\n=== Line Items ===")
		for _, item := range session.LineItems.Data {
			log.Printf("\nItem Description: %s", item.Description)
			log.Printf("Quantity: %d", item.Quantity)
			log.Printf("Unit Amount: %d", item.Price.UnitAmount)
			log.Printf("Product ID: %s", item.Price.Product.ID)

			log.Printf("Item Metadata:")
			for key, value := range item.Price.Product.Metadata {
				log.Printf("  %s: %s", key, value)
			}
		}

		// Print customer details
		if session.Customer != nil {
			log.Printf("\n=== Customer Details ===")
			log.Printf("Name: %s", session.Customer.Name)
			log.Printf("Email: %s", session.Customer.Email)
			log.Printf("Customer Metadata:")
			for key, value := range session.Customer.Metadata {
				log.Printf("  %s: %s", key, value)
			}
		}
	}

	return c.Status(200).JSON(fiber.Map{"received": true})
}
