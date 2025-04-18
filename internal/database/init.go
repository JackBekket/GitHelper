package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(connectionString string) (*Handler, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	handler := Handler{
		DB: db,
	}
	return &handler, nil
}
