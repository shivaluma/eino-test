package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/shivaluma/eino-agent/config"
)

var logger zerolog.Logger

func init() {
	// Initialize a simple console logger for migrations
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger = zerolog.New(output).With().Timestamp().Logger()
}

// Migration represents a single database migration
type Migration struct {
	Version     int64
	Filename    string
	Content     string
	Checksum    string
	RollbackSQL string
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version       int64
	Filename      string
	Checksum      string
	AppliedAt     time.Time
	ExecutionTime int
	Success       bool
	ErrorMessage  string
}

// Migrator manages database migrations
type Migrator struct {
	db            *pgxpool.Pool
	migrationsDir string
	config        *config.Config
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *pgxpool.Pool, migrationsDir string, cfg *config.Config) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		config:        cfg,
	}
}

// parseMigrationFilename extracts version from migration filename
// Expected format: 001_20250108000001_initial_schema.sql
func parseMigrationFilename(filename string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)_.*\.sql$`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) != 2 {
		return 0, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version number in filename %s: %w", filename, err)
	}

	return version, nil
}

// calculateChecksum calculates SHA-256 checksum of migration content
func calculateChecksum(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// LoadMigrations loads all migration files from the migrations directory
func (m *Migrator) LoadMigrations() ([]*Migration, error) {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []*Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		version, err := parseMigrationFilename(file.Name())
		if err != nil {
			// Skip files that don't match migration format
			continue
		}

		content, err := os.ReadFile(filepath.Join(m.migrationsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migration := &Migration{
			Version:  version,
			Filename: file.Name(),
			Content:  string(content),
			Checksum: calculateChecksum(string(content)),
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// InitializeMigrationSystem creates the migration tracking infrastructure
func (m *Migrator) InitializeMigrationSystem(ctx context.Context) error {
	// Check if schema_migrations table exists
	var exists bool
	err := m.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'schema_migrations'
		)
	`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check if migrations table exists: %w", err)
	}

	if !exists {
		// Run the migration system setup
		systemMigrationPath := filepath.Join(m.migrationsDir, "000_migration_system.sql")
		content, err := os.ReadFile(systemMigrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration system file: %w", err)
		}

		_, err = m.db.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to initialize migration system: %w", err)
		}

		logger.Info().Msg("✓ Migration system initialized")
	}

	return nil
}

// GetAppliedMigrations returns list of applied migrations
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]*MigrationStatus, error) {
	rows, err := m.db.Query(ctx, `
		SELECT version, filename, checksum, applied_at, execution_time_ms, success, error_message
		FROM schema_migrations
		WHERE version > 0  -- Skip system migration
		ORDER BY version
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var statuses []*MigrationStatus
	for rows.Next() {
		var status MigrationStatus
		var errorMsg sql.NullString

		err := rows.Scan(
			&status.Version,
			&status.Filename,
			&status.Checksum,
			&status.AppliedAt,
			&status.ExecutionTime,
			&status.Success,
			&errorMsg,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration status: %w", err)
		}

		if errorMsg.Valid {
			status.ErrorMessage = errorMsg.String
		}

		statuses = append(statuses, &status)
	}

	return statuses, nil
}

// GetCurrentVersion returns the highest successfully applied migration version
func (m *Migrator) GetCurrentVersion(ctx context.Context) (int64, error) {
	var version sql.NullInt64
	err := m.db.QueryRow(ctx, `
		SELECT MAX(version) 
		FROM schema_migrations 
		WHERE success = true AND version > 0
	`).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("failed to get current migration version: %w", err)
	}

	if !version.Valid {
		return 0, nil // No migrations applied
	}

	return version.Int64, nil
}

// ValidateMigration validates a migration hasn't been modified
func (m *Migrator) ValidateMigration(ctx context.Context, migration *Migration) error {
	var storedChecksum string
	err := m.db.QueryRow(ctx, `
		SELECT checksum FROM schema_migrations 
		WHERE version = $1 AND success = true
	`, migration.Version).Scan(&storedChecksum)

	if err == sql.ErrNoRows {
		return nil // Migration hasn't been applied
	}
	if err != nil {
		return fmt.Errorf("failed to validate migration checksum: %w", err)
	}

	if storedChecksum != migration.Checksum {
		return fmt.Errorf("migration %d has been modified (checksum mismatch)", migration.Version)
	}

	return nil
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(ctx context.Context, migration *Migration) error {
	startTime := time.Now()

	// Start transaction for the migration
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute the migration
	_, err = tx.Exec(ctx, migration.Content)
	executionTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		// Record the failed migration
		recordErr := m.recordMigrationExecution(ctx, migration, executionTime, false, err.Error())
		if recordErr != nil {
			logger.Warn().Err(recordErr).Msg("Failed to record migration failure")
		}
		return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
	}

	// Record successful migration
	err = m.recordMigrationExecution(ctx, migration, executionTime, true, "")
	if err != nil {
		return fmt.Errorf("failed to record migration success: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	logger.Info().
		Int64("version", migration.Version).
		Str("filename", migration.Filename).
		Float64("duration_seconds", time.Since(startTime).Seconds()).
		Msg("✓ Applied migration")

	return nil
}

// recordMigrationExecution records migration execution in schema_migrations table
func (m *Migrator) recordMigrationExecution(ctx context.Context, migration *Migration, executionTime int, success bool, errorMsg string) error {
	_, err := m.db.Exec(ctx, `
		INSERT INTO schema_migrations (version, filename, checksum, applied_at, execution_time_ms, success, error_message)
		VALUES ($1, $2, $3, NOW(), $4, $5, $6)
		ON CONFLICT (version) DO UPDATE SET
			filename = EXCLUDED.filename,
			checksum = EXCLUDED.checksum,
			applied_at = EXCLUDED.applied_at,
			execution_time_ms = EXCLUDED.execution_time_ms,
			success = EXCLUDED.success,
			error_message = EXCLUDED.error_message
	`, migration.Version, migration.Filename, migration.Checksum, executionTime, success, nullString(errorMsg))

	return err
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	// Initialize migration system if needed
	if err := m.InitializeMigrationSystem(ctx); err != nil {
		return err
	}

	// Load all migrations
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		logger.Info().Msg("No migrations found")
		return nil
	}

	// Get applied migrations for validation
	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Create a map for quick lookup
	appliedMap := make(map[int64]*MigrationStatus)
	for _, applied := range appliedMigrations {
		appliedMap[applied.Version] = applied
	}

	pendingCount := 0
	for _, migration := range migrations {
		// Skip system migration (version 0)
		if migration.Version == 0 {
			continue
		}

		// Check if migration is already applied
		if applied, exists := appliedMap[migration.Version]; exists {
			if applied.Success {
				// Validate migration hasn't been modified
				if err := m.ValidateMigration(ctx, migration); err != nil {
					return err
				}
				continue // Skip already applied migrations
			} else {
				logger.Warn().
					Int64("version", migration.Version).
					Str("error", applied.ErrorMessage).
					Msg("⚠ Migration previously failed")
				logger.Info().
					Int64("version", migration.Version).
					Str("filename", migration.Filename).
					Msg("Retrying migration")
			}
		}

		// Apply migration
		if err := m.ApplyMigration(ctx, migration); err != nil {
			return err
		}
		pendingCount++
	}

	if pendingCount == 0 {
		logger.Info().Msg("✓ Database is up to date")
	} else {
		logger.Info().Int("count", pendingCount).Msg("✓ Applied migrations")
	}

	return nil
}

// Status shows current migration status
func (m *Migrator) Status(ctx context.Context) error {
	// Initialize migration system if needed
	if err := m.InitializeMigrationSystem(ctx); err != nil {
		return err
	}

	// Get current version
	currentVersion, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	// Load all migrations
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	logger.Info().
		Int64("current_version", currentVersion).
		Int("total_migrations", len(migrations)-1).
		Msg("Migration status")

	if len(appliedMigrations) > 0 {
		logger.Info().Msg("Applied migrations:")
		for _, applied := range appliedMigrations {
			status := "✓"
			if !applied.Success {
				status = "✗"
			}
			logger.Info().
				Str("status", status).
				Int64("version", applied.Version).
				Str("filename", applied.Filename).
				Str("applied_at", applied.AppliedAt.Format("2006-01-02 15:04:05")).
				Msg("")
		}
	}

	// Show pending migrations
	appliedMap := make(map[int64]bool)
	for _, applied := range appliedMigrations {
		if applied.Success {
			appliedMap[applied.Version] = true
		}
	}

	var pendingMigrations []*Migration
	for _, migration := range migrations {
		if migration.Version == 0 { // Skip system migration
			continue
		}
		if !appliedMap[migration.Version] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) > 0 {
		logger.Info().Msg("Pending migrations:")
		for _, migration := range pendingMigrations {
			logger.Info().
				Str("status", "○").
				Int64("version", migration.Version).
				Str("filename", migration.Filename).
				Msg("")
		}
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback(ctx context.Context) error {
	// Get current version
	currentVersion, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	if currentVersion == 0 {
		logger.Info().Msg("No migrations to rollback")
		return nil
	}

	// Get the migration to rollback
	var rollbackSQL sql.NullString
	var filename string
	err = m.db.QueryRow(ctx, `
		SELECT filename, rollback_sql 
		FROM schema_migrations 
		WHERE version = $1 AND success = true
	`, currentVersion).Scan(&filename, &rollbackSQL)

	if err != nil {
		return fmt.Errorf("failed to get rollback information for migration %d: %w", currentVersion, err)
	}

	if !rollbackSQL.Valid || rollbackSQL.String == "" {
		return fmt.Errorf("migration %d (%s) does not have rollback SQL", currentVersion, filename)
	}

	logger.Info().
		Int64("version", currentVersion).
		Str("filename", filename).
		Msg("Rolling back migration")

	// Execute rollback in transaction
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start rollback transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, rollbackSQL.String)
	if err != nil {
		return fmt.Errorf("failed to execute rollback for migration %d: %w", currentVersion, err)
	}

	// Mark migration as rolled back
	_, err = tx.Exec(ctx, `
		UPDATE schema_migrations 
		SET success = false, error_message = 'Rolled back at ' || NOW() 
		WHERE version = $1
	`, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to mark migration as rolled back: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	logger.Info().
		Int64("version", currentVersion).
		Msg("✓ Successfully rolled back migration")
	return nil
}

// RollbackTo rolls back to a specific migration version
func (m *Migrator) RollbackTo(ctx context.Context, targetVersion int64) error {
	currentVersion, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	if targetVersion >= currentVersion {
		logger.Info().
			Int64("target_version", targetVersion).
			Int64("current_version", currentVersion).
			Msg("Target version is not lower than current version")
		return nil
	}

	// Get all migrations to rollback (in reverse order)
	rows, err := m.db.Query(ctx, `
		SELECT version, filename, rollback_sql
		FROM schema_migrations
		WHERE version > $1 AND version <= $2 AND success = true
		ORDER BY version DESC
	`, targetVersion, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to query migrations for rollback: %w", err)
	}
	defer rows.Close()

	var migrationsToRollback []struct {
		Version     int64
		Filename    string
		RollbackSQL string
	}

	for rows.Next() {
		var m struct {
			Version     int64
			Filename    string
			RollbackSQL sql.NullString
		}

		err := rows.Scan(&m.Version, &m.Filename, &m.RollbackSQL)
		if err != nil {
			return fmt.Errorf("failed to scan migration for rollback: %w", err)
		}

		if !m.RollbackSQL.Valid || m.RollbackSQL.String == "" {
			return fmt.Errorf("migration %d (%s) does not have rollback SQL", m.Version, m.Filename)
		}

		migrationsToRollback = append(migrationsToRollback, struct {
			Version     int64
			Filename    string
			RollbackSQL string
		}{
			Version:     m.Version,
			Filename:    m.Filename,
			RollbackSQL: m.RollbackSQL.String,
		})
	}

	if len(migrationsToRollback) == 0 {
		logger.Info().Msg("No migrations to rollback")
		return nil
	}

	logger.Info().
		Int("count", len(migrationsToRollback)).
		Int64("target_version", targetVersion).
		Msg("Rolling back migrations")

	// Execute rollbacks in reverse order
	for _, migration := range migrationsToRollback {
		logger.Info().
			Int64("version", migration.Version).
			Str("filename", migration.Filename).
			Msg("Rolling back migration")

		tx, err := m.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start rollback transaction for migration %d: %w", migration.Version, err)
		}

		_, err = tx.Exec(ctx, migration.RollbackSQL)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute rollback for migration %d: %w", migration.Version, err)
		}

		// Mark migration as rolled back
		_, err = tx.Exec(ctx, `
			UPDATE schema_migrations 
			SET success = false, error_message = 'Rolled back at ' || NOW()
			WHERE version = $1
		`, migration.Version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to mark migration %d as rolled back: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback transaction for migration %d: %w", migration.Version, err)
		}

		logger.Info().
			Int64("version", migration.Version).
			Msg("✓ Rolled back migration")
	}

	return nil
}

// Validate validates all applied migrations against their files
func (m *Migrator) Validate(ctx context.Context) error {
	// Load all migrations
	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	// Create migration map for quick lookup
	migrationMap := make(map[int64]*Migration)
	for _, migration := range migrations {
		migrationMap[migration.Version] = migration
	}

	logger.Info().Msg("Validating migration checksums...")

	var errors []string
	for _, applied := range appliedMigrations {
		if !applied.Success {
			continue // Skip failed migrations
		}

		migration, exists := migrationMap[applied.Version]
		if !exists {
			errors = append(errors, fmt.Sprintf("Applied migration %d (%s) not found in migrations directory", applied.Version, applied.Filename))
			continue
		}

		if migration.Checksum != applied.Checksum {
			errors = append(errors, fmt.Sprintf("Migration %d (%s) has been modified (checksum mismatch)", applied.Version, applied.Filename))
		}
	}

	if len(errors) > 0 {
		logger.Error().Msg("❌ Migration validation failed:")
		for _, err := range errors {
			logger.Error().Str("error", err).Msg("•")
		}
		return fmt.Errorf("migration validation failed with %d errors", len(errors))
	}

	logger.Info().Msg("✓ All migrations validated successfully")
	return nil
}

// Reset drops all tables and reapplies all migrations (DANGEROUS!)
func (m *Migrator) Reset(ctx context.Context, confirmed bool) error {
	if !confirmed {
		return fmt.Errorf("reset operation requires explicit confirmation. This will DROP ALL TABLES")
	}

	logger.Warn().Msg("⚠ RESETTING DATABASE - This will drop all tables and data!")

	// Drop all tables
	_, err := m.db.Exec(ctx, `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO public;
	`)
	if err != nil {
		return fmt.Errorf("failed to reset database: %w", err)
	}

	logger.Info().Msg("✓ Database reset complete")
	logger.Info().Msg("Reapplying all migrations...")

	// Reapply all migrations
	return m.Migrate(ctx)
}

// nullString returns sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
