package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomOTP() string {
	// We want values from 1 to 999998 inclusive → total 999998 values
	max := big.NewInt(999998)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return ""
	}

	// Shift range from 0–999997 → 1–999998
	value := n.Int64() + 1

	// strNumber := strconv.FormatInt(value, 10)
	fmt.Printf("Generated OTP: %06d\n", value)
	// formattedNumber := strNumber[:3] + "-" + strNumber[3:6]
	return fmt.Sprintf("%06d", value)
}
