package utils

import (
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
