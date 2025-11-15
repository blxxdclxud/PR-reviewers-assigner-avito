package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/config"
	_ "github.com/lib/pq"
)

const (
	MaxOpenConnections    = 25
	MaxIdleConnections    = 5
	MaxConnectionLifetime = 3 * time.Minute
	MaxConnectionIdleTime = 10 * time.Minute
)

// NewDB creates and sets up database connection pool
func NewDB(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)

	// Create DB connection pool
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	// Set up connection pool
	db.SetMaxIdleConns(MaxIdleConnections)
	db.SetMaxOpenConns(MaxOpenConnections)
	db.SetConnMaxIdleTime(MaxConnectionIdleTime)
	db.SetConnMaxLifetime(MaxConnectionLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	return db, nil
}

// CloseDB closes DB connection
func CloseDB(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
