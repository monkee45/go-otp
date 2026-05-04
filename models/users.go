package models

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type User struct {
	ID        int
	Name      string
	Email     string
	Otp       []byte
	OtpExpiry time.Time
	CreatedAt time.Time
}

type UserService struct {
	DB     *sql.DB
	Logger *slog.Logger
}

type ErrEmailTaken struct {
	Email string
}

type ErrUserNotExist struct {
	Email string
}

func (err ErrEmailTaken) Error() string {
	return fmt.Sprintf("email address %q is already taken", err.Email)
}

func (err ErrUserNotExist) Error() string {
	return fmt.Sprintf("User address %q does not exist", err.Email)
}

func (userService *UserService) Create(user *User) error {
	email := strings.ToLower(user.Email)
	existingUser, err := userService.FindByEmail(email)
	if err != nil {
		if err != sql.ErrNoRows {
			return ErrUserNotExist{Email: email}
		}
	}
	if existingUser != nil {
		return ErrEmailTaken{Email: email}
	}

	row := userService.DB.QueryRow(`
		INSERT INTO users (name, email, otpExpiry)
		VALUES ($1, $2, $3) RETURNING id, created_at`,
		user.Name, user.Email, time.Now().AddDate(0, 0, -1))

	err = row.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	userService.Logger.Info("Create New User: ", "Email", user.Email)
	return nil
}

func (userService *UserService) FindByEmail(email string) (*User, error) {
	email = strings.ToLower(email)
	user := &User{
		Email: email,
	}
	var hexOTP string
	// Check if record exists
	row := userService.DB.QueryRow(`
		SELECT id, name, email, COALESCE(otp, ''), otpexpiry, created_at FROM users WHERE email = $1`, email)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&hexOTP,
		&user.OtpExpiry,
		&user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	user.Otp, _ = hex.DecodeString(hexOTP)

	return user, nil
}

func (userService *UserService) UpdateOTP(email string, otp []byte) error {
	hexString := hex.EncodeToString(otp)
	otpExpiry := time.Now().Add(5 * time.Minute)

	_, err := userService.DB.Exec(`
		UPDATE users 
		SET otp = $2,
		OtpExpiry = $3
		WHERE email = $1;`, email, hexString, otpExpiry)

	if err != nil {
		return fmt.Errorf("update One Time Password Failed: %w", err)
	}
	return nil
}
