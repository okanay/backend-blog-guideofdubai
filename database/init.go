package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Init(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	fmt.Println("Opening the database connection")

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to the database")

	return db, nil
}

func Close(db *sql.DB) error {
	err := db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	fmt.Println("Closing the database connection")
	return nil
}
