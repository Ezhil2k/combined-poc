package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() error {
	connStr := "host=localhost port=5433 user=gobackend_user password=strong_secure_password dbname=verity sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

