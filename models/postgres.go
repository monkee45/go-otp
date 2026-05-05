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
		return nil, fmt.Errorf("Error opening SQL DB connection: %v.  Error: %v", psqlInfo, err)
	}
	err = sqldb.Ping()
	if err != nil {
		return nil, fmt.Errorf("Could not ping database: %v.  Error: %v", psqlInfo, err)
	}

	return sqldb, err
}

func (s *Services) Close() {
	s.db.Close()
	log.Printf("Database closed")
}
