package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
)

func Connect() (*pgx.Conn, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://user:password@localhost:5432/dbname"
	}
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to database")
	return conn, nil
}
