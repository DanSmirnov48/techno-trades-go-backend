package utils

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomStr := make([]byte, length)
	for i := range randomStr {
		randomStr[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomStr)
}

// Generates a random integer with a specified number of digits
func GetRandomInt(size int) int64 {
	if size <= 0 {
		return 0
	}

	// Calculate the min and max possible values for the specified size
	min := intPow(10, size-1)
	max := intPow(10, size) - 1

	// Initialize the random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random integer within the range [min, max]
	return int64(rand.Intn(max-min+1) + min)
}

// intPow calculates the power of base^exponent for integers
func intPow(base, exponent int) int {
	result := 1
	for i := 0; i < exponent; i++ {
		result *= base
	}
	return result
}

// UUID PARSER
func ParseUUID(input string) (*uuid.UUID, *ErrorResponse) {
	uuidVal, err := uuid.Parse(input)
	if err != nil {
		errData := RequestErr(ERR_INVALID_VALUE, "Invalid UUID")
		return nil, &errData
	}
	return &uuidVal, nil
}
