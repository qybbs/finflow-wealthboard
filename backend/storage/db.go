package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func InitDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "finflow")
	sslmode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL database and initialized tables.")
	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(100) PRIMARY KEY,
			date DATE NOT NULL,
			type VARCHAR(50) NOT NULL,
			category VARCHAR(100) NOT NULL,
			amount DOUBLE PRECISION NOT NULL,
			description TEXT,
			method VARCHAR(50) NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS assets (
			id VARCHAR(100) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			code VARCHAR(50) NOT NULL UNIQUE,
			quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
			price_per_unit DOUBLE PRECISION NOT NULL DEFAULT 0,
			average_price DOUBLE PRECISION NOT NULL DEFAULT 0,
			current_price DOUBLE PRECISION NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS journal_entries (
			id VARCHAR(100) PRIMARY KEY,
			date DATE NOT NULL,
			asset_id VARCHAR(100) NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
			transaction_id VARCHAR(100) REFERENCES transactions(id) ON DELETE SET NULL,
			type VARCHAR(50) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			price_per_unit DOUBLE PRECISION NOT NULL,
			fee DOUBLE PRECISION NOT NULL DEFAULT 0,
			total DOUBLE PRECISION NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS budgets (
			category VARCHAR(100) PRIMARY KEY,
			limit_amount DOUBLE PRECISION NOT NULL,
			interval VARCHAR(50) NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %v\nQuery: %s", err, query)
		}
	}

	return nil
}
