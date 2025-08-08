package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/shivaluma/eino-agent/config"
	"github.com/shivaluma/eino-agent/internal/migrations"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Parse command line arguments
	var (
		command = flag.String("command", "migrate", "Command to run: migrate, status, rollback, rollback-to, validate, reset, generate")
		version = flag.Int64("version", 0, "Target version for rollback-to command")
		confirm = flag.Bool("confirm", false, "Confirm destructive operations like reset")
		name    = flag.String("name", "", "Name for new migration (required for generate command)")
	)
	flag.Parse()

	// Handle generate command early (doesn't need database connection)
	if *command == "generate" {
		if *name == "" {
			log.Fatal("Migration name is required for generate command. Use -name=your_migration_name")
		}
		if err := generateMigration(*name); err != nil {
			log.Fatalf("Failed to generate migration: %v", err)
		}
		return
	}

	// Initialize configuration
	cfg := config.Load()

	// Build database URL
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database)

	// Connect to database
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize migrator
	migrator := migrations.NewMigrator(db, "migrations", cfg)

	// Execute command
	switch *command {
	case "migrate":
		if err := migrator.Migrate(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✓ Migrations completed successfully")

	case "status":
		if err := migrator.Status(ctx); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "rollback":
		if err := migrator.Rollback(ctx); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

	case "rollback-to":
		if *version <= 0 {
			log.Fatal("Version must be specified and greater than 0 for rollback-to command")
		}
		if err := migrator.RollbackTo(ctx, *version); err != nil {
			log.Fatalf("Rollback to version %d failed: %v", *version, err)
		}

	case "validate":
		if err := migrator.Validate(ctx); err != nil {
			log.Fatalf("Migration validation failed: %v", err)
		}

	case "reset":
		if !*confirm {
			fmt.Println("⚠ WARNING: This will DROP ALL TABLES and reapply all migrations!")
			fmt.Println("To confirm, add the -confirm flag:")
			fmt.Printf("  go run cmd/migrate/main.go -command=reset -confirm\n")
			os.Exit(1)
		}
		if err := migrator.Reset(ctx, true); err != nil {
			log.Fatalf("Database reset failed: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		fmt.Fprintf(os.Stderr, "Available commands: migrate, status, rollback, rollback-to, validate, reset, generate\n")
		flag.Usage()
		os.Exit(1)
	}
}

// generateMigration creates a new migration file with proper naming convention
func generateMigration(name string) error {
	// Get current migrations to determine next version number
	migrations, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to list existing migrations: %w", err)
	}

	// Find the highest version number
	maxVersion := int64(0)
	for _, migration := range migrations {
		basename := filepath.Base(migration)
		if strings.HasPrefix(basename, "000_") {
			continue // Skip system migration
		}

		// Extract version number from filename (format: 001_timestamp_name.sql)
		parts := strings.Split(basename, "_")
		if len(parts) >= 1 {
			if version, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				if version > maxVersion {
					maxVersion = version
				}
			}
		}
	}

	// Generate next version number
	nextVersion := maxVersion + 1

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Clean up migration name (replace spaces with underscores, lowercase)
	cleanName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	// Generate filename
	filename := fmt.Sprintf("%03d_%s_%s.sql", nextVersion, timestamp, cleanName)
	filepath := filepath.Join("migrations", filename)

	// Generate migration template
	template := `-- Migration: ` + name + `
-- Created: ` + time.Now().Format("2006-01-02 15:04:05") + `
-- Version: ` + fmt.Sprintf("%d", nextVersion) + `

-- Add your SQL statements here
-- Example:
-- CREATE TABLE example (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- Rollback SQL (optional - add rollback statements as comments)
-- This migration does not have automatic rollback
-- To rollback manually, run the reverse operations:
-- DROP TABLE IF EXISTS example;
`

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll("migrations", 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Write the migration file
	if err := os.WriteFile(filepath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	fmt.Printf("✓ Generated migration file: %s\n", filename)
	fmt.Printf("✓ Migration version: %d\n", nextVersion)
	fmt.Printf("✓ File path: %s\n", filepath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit the migration file to add your SQL statements")
	fmt.Println("2. Run 'make db-migrate' to apply the migration")
	fmt.Println("3. Run 'make db-migrate-status' to verify the migration")

	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Migration CLI Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  migrate      - Run all pending migrations (default)\n")
		fmt.Fprintf(os.Stderr, "  status       - Show current migration status\n")
		fmt.Fprintf(os.Stderr, "  rollback     - Rollback the last migration\n")
		fmt.Fprintf(os.Stderr, "  rollback-to  - Rollback to a specific migration version\n")
		fmt.Fprintf(os.Stderr, "  validate     - Validate all migration checksums\n")
		fmt.Fprintf(os.Stderr, "  reset        - DROP ALL TABLES and reapply migrations (DANGEROUS)\n")
		fmt.Fprintf(os.Stderr, "  generate     - Generate a new migration file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                                     # Run pending migrations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=status                     # Show migration status\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=rollback                   # Rollback last migration\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=rollback-to -version=2     # Rollback to version 2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=validate                   # Validate migrations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=reset -confirm             # Reset database\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -command=generate -name=\"add_users\" # Generate new migration\n", os.Args[0])
	}
}
