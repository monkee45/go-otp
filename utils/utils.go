package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
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
	// formattedNumber := strNumber[:3] + "-" + strNumber[3:6]
	return fmt.Sprintf("%06d", value)
}

func HashOTP(otp string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
}

func CompareOTP(hash []byte, otp string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(otp)) == nil
}
