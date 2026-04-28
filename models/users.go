package models

import (
	"database/sql"
	"fmt"
	"log"
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

type UserDBService struct {
	DB *sql.DB
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

func (udb *UserDBService) Create(user *User) error {
	email := strings.ToLower(user.Email)
	// Check if email is already taken
	row := udb.DB.QueryRow(`
		SELECT id FROM users WHERE email = $1`, email)
	var id int
	err := row.Scan(&id)
	if err != sql.ErrNoRows {
		if err == nil {
			return ErrEmailTaken{Email: email}
		} else {
			return err
		}
	}

	row = udb.DB.QueryRow(`
		INSERT INTO users (name, email)
		VALUES ($1, $2) RETURNING id, created_at`,
		user.Name, user.Email)

	err = row.Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (udb *UserDBService) FindByEmail(email string) (*User, error) {
	log.Printf("--> models.FindByEmail\n")
	email = strings.ToLower(email)
	user := &User{
		Email: email,
	}
	// Check if record exists
	row := udb.DB.QueryRow(`
		SELECT id, name, email, otp, otpexpiry, created_at FROM users WHERE email = $1`, email)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Otp,
		&user.OtpExpiry,
		&user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotExist{Email: email}
		}
		return nil, err
	}
	return user, nil
}

func (udb *UserDBService) UpdateOTP(email string, otp []byte) error {
	// hexString := hex.EncodeToString(otp)
	otpExpiry := time.Now().Add(5 * time.Minute)
	_, err := udb.DB.Exec(`
		UPDATE users 
		SET otp = $2,
		OtpExpiry = $3
		WHERE email = $1;`, email, otp, otpExpiry)

	if err != nil {
		return fmt.Errorf("update One Time Password Failed: %w", err)
	}
	return nil
}
