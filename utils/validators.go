package utils

import (
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validates if a date has a correct format (ISO8601)
func DateValidator(fl validator.FieldLevel) bool {
	inputTimeString := fl.Field().String()
	_, err := time.Parse(time.RFC3339, inputTimeString)
	return err == nil
}

func ValidateUUID(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	_, err := uuid.Parse(value)
	return err == nil
}

func ValidateDiscountedPrice(fl validator.FieldLevel) bool {
	isDiscountedField := fl.Parent().FieldByName("IsDiscounted")
	if !isDiscountedField.IsValid() || isDiscountedField.Kind() != reflect.Bool {
		return false
	}

	isDiscounted := isDiscountedField.Bool()
	discountedPriceField := fl.Parent().FieldByName("DiscountedPrice")
	if !discountedPriceField.IsValid() || discountedPriceField.Kind() != reflect.Float64 {
		return false
	}

	discountedPrice := discountedPriceField.Float()
	if isDiscounted && discountedPrice <= 0 {
		return false
	}

	if !isDiscounted && discountedPrice != 0 {
		return false
	}

	return true
}
