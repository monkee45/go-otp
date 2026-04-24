package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Services struct {
	db *sql.DB
}

func Open(psqlInfo string) (*sql.DB, error) {
	sqldb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("opening SQL DB connection: %w", err)
	}
	err = sqldb.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not contact database, %w", err)
	} else {
		log.Println("able to contact the database")
	}

	return sqldb, err
}

func (s *Services) Close() {
	s.db.Close()
	log.Printf("Database closed!")
}
