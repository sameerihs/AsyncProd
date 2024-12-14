package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Create connection config
	connConfig, err := pgx.ParseConfig(fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	))
	if err != nil {
		log.Fatalf("Error creating database config: %v", err)
	}

	// pgx config to a db/sql
	DB = stdlib.OpenDB(*connConfig)


	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully")
	err = createTables()
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)
}

// createTables creates necessary tables if they don't exist
func createTables() error {
	// for users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// for products table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			product_description TEXT,
			product_price DECIMAL(10,2) NOT NULL,
			product_images TEXT[],
			compressed_product_images TEXT[],
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating products table: %v", err)
	}

	// insert a test user if no users exist
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking users: %v", err)
	}

	if count == 0 {
		_, err = DB.Exec("INSERT INTO users (username) VALUES ($1)", "testuser")
		if err != nil {
			return fmt.Errorf("error inserting test user: %v", err)
		}
		log.Println("Inserted test user")
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}