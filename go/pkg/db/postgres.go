package db

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool
var initOnce sync.Once
var initError error

// InitDB initializes the database connection pool (singleton)
func InitDB() error {
	initOnce.Do(func() {
		supabaseURL := os.Getenv("SUPABASE_URL")
		dbPassword := os.Getenv("DB_PASSWORD")

		if supabaseURL == "" {
			initError = fmt.Errorf("SUPABASE_URL must be set")
			return
		}

		if dbPassword == "" {
			initError = fmt.Errorf("DB_PASSWORD must be set")
			return
		}

		// Extract project ref from Supabase URL
		// Format: https://xxx.supabase.co -> xxx
		projectRef := extractProjectRef(supabaseURL)
		
		if projectRef == "" {
			initError = fmt.Errorf("invalid SUPABASE_URL format: %s", supabaseURL)
			return
		}

		// URL encode the password to handle special characters
		encodedPassword := url.QueryEscape(dbPassword)

		// Construct PostgreSQL connection string using Supabase Pooler
		// Supabase uses connection pooling with format:
		// postgresql://postgres.PROJECT_REF:PASSWORD@aws-1-ap-south-1.pooler.supabase.com:6543/postgres
		connString := fmt.Sprintf(
			"postgresql://postgres.%s:%s@aws-1-ap-south-1.pooler.supabase.com:6543/postgres?sslmode=require",
			projectRef,
			encodedPassword,
		)

		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			initError = fmt.Errorf("failed to parse connection string: %w", err)
			return
		}

		// Set connection timeout
		config.ConnConfig.ConnectTimeout = 10 * time.Second
		
		// Disable prepared statement cache to avoid issues in Lambda
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			initError = fmt.Errorf("failed to create connection pool: %w", err)
			return
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			initError = fmt.Errorf("failed to ping database: %w", err)
			return
		}
	})

	return initError
}

// extractProjectRef extracts the project reference from Supabase URL
// https://xxx.supabase.co -> xxx
func extractProjectRef(urlStr string) string {
	// Remove https:// prefix
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")
	
	// Split by .supabase.co
	parts := strings.Split(urlStr, ".supabase.co")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	
	return ""
}

// GetPool returns the database connection pool
func GetPool() *pgxpool.Pool {
	return pool
}

// Close closes the database connection pool
func Close() {
	if pool != nil {
		pool.Close()
	}
}
