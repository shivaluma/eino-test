-- Migration tracking system - This must be the first migration
-- Creates the infrastructure to track all subsequent migrations

-- Create schema_migrations table to track applied migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL, -- SHA-256 checksum of the migration file
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    rollback_sql TEXT -- Optional rollback SQL for reversible migrations
);

-- Create index for efficient querying
CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_success ON schema_migrations(success);

-- Insert this migration as the first tracked migration
INSERT INTO schema_migrations (version, filename, checksum, applied_at, execution_time_ms, success)
VALUES (0, '000_migration_system.sql', '00000000000000000000000000000000', NOW(), 0, true)
ON CONFLICT (version) DO NOTHING;

-- Create function to validate migration checksums
CREATE OR REPLACE FUNCTION validate_migration_checksum(
    migration_version BIGINT,
    expected_checksum VARCHAR(64)
) RETURNS BOOLEAN AS $$
DECLARE
    stored_checksum VARCHAR(64);
BEGIN
    SELECT checksum INTO stored_checksum 
    FROM schema_migrations 
    WHERE version = migration_version AND success = true;
    
    IF NOT FOUND THEN
        RETURN true; -- Migration hasn't been applied yet
    END IF;
    
    RETURN stored_checksum = expected_checksum;
END;
$$ LANGUAGE plpgsql;

-- Create function to get the current migration version
CREATE OR REPLACE FUNCTION get_current_migration_version() RETURNS BIGINT AS $$
DECLARE
    current_version BIGINT;
BEGIN
    SELECT COALESCE(MAX(version), -1) INTO current_version
    FROM schema_migrations
    WHERE success = true;
    
    RETURN current_version;
END;
$$ LANGUAGE plpgsql;

-- Create function to record migration execution
CREATE OR REPLACE FUNCTION record_migration_execution(
    migration_version BIGINT,
    migration_filename VARCHAR(255),
    migration_checksum VARCHAR(64),
    execution_time INTEGER,
    migration_success BOOLEAN,
    error_msg TEXT DEFAULT NULL,
    rollback_query TEXT DEFAULT NULL
) RETURNS VOID AS $$
BEGIN
    INSERT INTO schema_migrations (
        version, filename, checksum, applied_at, 
        execution_time_ms, success, error_message, rollback_sql
    ) VALUES (
        migration_version, migration_filename, migration_checksum, NOW(),
        execution_time, migration_success, error_msg, rollback_query
    )
    ON CONFLICT (version) DO UPDATE SET
        filename = EXCLUDED.filename,
        checksum = EXCLUDED.checksum,
        applied_at = EXCLUDED.applied_at,
        execution_time_ms = EXCLUDED.execution_time_ms,
        success = EXCLUDED.success,
        error_message = EXCLUDED.error_message,
        rollback_sql = EXCLUDED.rollback_sql;
END;
$$ LANGUAGE plpgsql;