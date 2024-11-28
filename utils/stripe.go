package utils

import (
	"fmt"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/stripe/stripe-go/v81"
)

var (
	successURL = fmt.Sprintf("%s/checkout-success?session_id={CHECKOUT_SESSION_ID}", config.GetConfig().FrontendURL)
	SuccessURL = stripe.String(successURL)

	cancelURL = fmt.Sprintf("%s/cart", config.GetConfig().FrontendURL)
	CancelURL = stripe.String(cancelURL)
)

func GetStripeAllowedCountries() *stripe.CheckoutSessionShippingAddressCollectionParams {
	return &stripe.CheckoutSessionShippingAddressCollectionParams{
		AllowedCountries: stripe.StringSlice([]string{"GB"}),
	}
}

func GetStripeShippingOptions() []*stripe.CheckoutSessionShippingOptionParams {
	return []*stripe.CheckoutSessionShippingOptionParams{
		{
			ShippingRateData: &stripe.CheckoutSessionShippingOptionShippingRateDataParams{
				Type: stripe.String("fixed_amount"),
				FixedAmount: &stripe.CheckoutSessionShippingOptionShippingRateDataFixedAmountParams{
					Amount:   stripe.Int64(0),
					Currency: stripe.String("gbp"),
				},
				DisplayName: stripe.String("Free shipping"),
				DeliveryEstimate: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateParams{
					Minimum: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateMinimumParams{
						Unit:  stripe.String("business_day"),
						Value: stripe.Int64(5),
					},
					Maximum: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateMaximumParams{
						Unit:  stripe.String("business_day"),
						Value: stripe.Int64(7),
					},
				},
			},
		},
		{
			ShippingRateData: &stripe.CheckoutSessionShippingOptionShippingRateDataParams{
				Type: stripe.String("fixed_amount"),
				FixedAmount: &stripe.CheckoutSessionShippingOptionShippingRateDataFixedAmountParams{
					Amount:   stripe.Int64(1500),
					Currency: stripe.String("gbp"),
				},
				DisplayName: stripe.String("Next day air"),
				DeliveryEstimate: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateParams{
					Minimum: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateMinimumParams{
						Unit:  stripe.String("business_day"),
						Value: stripe.Int64(1),
					},
					Maximum: &stripe.CheckoutSessionShippingOptionShippingRateDataDeliveryEstimateMaximumParams{
						Unit:  stripe.String("business_day"),
						Value: stripe.Int64(1),
					},
				},
			},
		},
	}
}
