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
	CreatedAt time.Time
}

type UserDBService struct {
	DB *sql.DB
}

type ErrEmailTaken struct {
	Email string
}

func (err ErrEmailTaken) Error() string {
	return fmt.Sprintf("email address %q is already taken", err.Email)
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
	log.Println("result of row.Scan(&user): ", user, err)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}
