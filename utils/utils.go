package utils

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

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

// create anew logger that will log to a given logfile
func NewLogger(logfile string) *slog.Logger {
	const layout = "2006-01-02"
	t := time.Now()
	fname := logfile + "-" + t.Format(layout) + ".log"
	// Open (or create) a log file
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	// Create a handler that writes to the file
	handler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo, // minimum log level
	})

	// Create a logger using that handler
	logger := slog.New(handler)

	return logger
}
