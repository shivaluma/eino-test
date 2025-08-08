# Development Guide

This guide covers development setup and workflows for the AI Food Agent project.

## Quick Start

### First-time Setup
```bash
# Clone the repository and setup environment
make setup

# Start development server with live reload
make dev
```

## Development Commands

### Primary Development
- `make dev` - Start development server with live reload (recommended)
- `make server` - Run server directly without live reload
- `make air` - Start Air live reload explicitly

### Building
- `make build` - Build application for current platform
- `make build-prod` - Build optimized binary for production (Linux)

### Testing
- `make test` - Run all tests
- `make test-verbose` - Run tests with verbose output
- `make test-coverage` - Generate test coverage report

### Code Quality
- `make fmt` - Format all Go code
- `make vet` - Run Go vet for static analysis
- `make lint` - Run golangci-lint (requires installation)

### Database Management
- `make db-migrate` - Run all pending database migrations
- `make db-migrate-status` - Show current migration status
- `make db-migrate-generate NAME="migration_name"` - Generate new migration file
- `make db-migrate-rollback` - Rollback the last migration
- `make db-migrate-rollback-to VERSION=X` - Rollback to specific version
- `make db-migrate-validate` - Validate migration checksums
- `make db-migrate-reset-confirmed` - Reset database (WARNING: destructive)
- `make db-reset` - Alias for db-migrate-reset (shows warning first)
- `make db-connect` - Connect to database via psql

### Docker Development
- `make docker-up` - Start all services (database + API)
- `make docker-down` - Stop all services
- `make docker-logs` - View logs from all services
- `make docker-rebuild` - Rebuild and restart services

### Dependencies
- `make deps` - Download Go dependencies
- `make tidy` - Clean up Go modules
- `make vendor` - Create vendor directory

### Cleanup
- `make clean` - Remove build artifacts and temporary files
- `make clean-all` - Remove all generated files including vendor

### Setup & Tools
- `make setup` - Initial development environment setup
- `make install-tools` - Install development tools (Air, golangci-lint)
- `make help` - Show all available commands

## Development Workflow

### Starting Development
1. **First time setup:**
   ```bash
   make setup
   ```

2. **Daily development:**
   ```bash
   make dev
   ```
   This starts the server with live reload. Any changes to `.go` files will automatically rebuild and restart the server.

### Code Changes
1. Make your changes to the Go files
2. Air will automatically detect changes and rebuild
3. Test your changes at `http://localhost:8888` (or your configured port)

### Running Tests
```bash
# Quick test run
make test

# Detailed test output
make test-verbose

# Generate coverage report
make test-coverage
# Open coverage.html in your browser to view coverage
```

### Database Changes
```bash
# Apply pending migrations
make db-migrate

# Check migration status
make db-migrate-status

# Create new migration
make db-migrate-generate NAME="add_new_feature"

# Rollback last migration
make db-migrate-rollback

# Reset database (careful - this drops all data!)
make db-reset

# Connect to database for manual queries
make db-connect
```

## Configuration

### Environment Variables
Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:
- `DB_*` - Database connection settings
- `JWT_*` - JWT token configuration  
- `SERVER_*` - Server settings
- `OPENAI_*` - AI integration settings
- `OAUTH_*` - OAuth provider configurations (GitHub, Google)
- `FRONTEND_URL` - Frontend URL for OAuth redirects

### Air Configuration
Live reload is configured in `.air.toml`. Key settings:
- Watches all `.go` files
- Excludes test files and vendor directories
- Builds to `./tmp/main`
- Automatically restarts on changes

## Debugging

### Server Logs
When using `make dev`, all server logs appear in the terminal with color-coded output from Air.

### Database Connection Issues
```bash
# Test database connection
make db-connect

# Check database logs in Docker
make docker-logs
```

### Build Issues
```bash
# Clean and rebuild
make clean
make build

# Check for formatting/vet issues
make fmt
make vet
```

## Tools Integration

### VS Code
Recommended extensions:
- Go extension
- REST Client (for API testing)

### Git Hooks
Consider adding pre-commit hooks:
```bash
#!/bin/sh
make fmt
make vet
make test
```

## Performance Tips

1. **Use Air for development** - Much faster than manual restarts
2. **Keep database running** - Use `make docker-up` to start PostgreSQL once
3. **Use connection pooling** - Already configured in the database layer
4. **Watch build logs** - Air shows compilation errors immediately

## Migration System

### Creating New Migrations
```bash
# Generate a new migration file
make db-migrate-generate NAME="add_user_preferences"
# This creates: migrations/003_20250809123456_add_user_preferences.sql

# Edit the generated file to add your SQL
# Then apply it:
make db-migrate
```

### Migration File Format
Generated migrations follow the pattern: `XXX_YYYYMMDDHHMMSS_name.sql`
- `XXX` = Sequential version (001, 002, 003...)
- `YYYYMMDDHHMMSS` = Timestamp
- `name` = Sanitized migration name

### Migration Safety
- All migrations run in transactions
- Checksums prevent tampering with applied migrations
- Auto-migration runs on server startup
- Rollback support for safe reversions

## OAuth Development

### Testing OAuth Locally
1. Set up GitHub/Google OAuth applications
2. Configure redirect URIs: `http://localhost:8888/api/v1/auth/oauth/github/callback`
3. Add credentials to `.env`:
   ```
   OAUTH_GITHUB_CLIENT_ID=your_github_client_id
   OAUTH_GITHUB_CLIENT_SECRET=your_github_client_secret
   OAUTH_GOOGLE_CLIENT_ID=your_google_client_id
   OAUTH_GOOGLE_CLIENT_SECRET=your_google_client_secret
   ```

### OAuth Flow Testing
```bash
# Get available providers
curl http://localhost:8888/api/v1/auth/oauth/providers

# Initiate OAuth (will redirect to provider)
curl http://localhost:8888/api/v1/auth/oauth/github/authorize

# Check linked accounts (requires auth)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/api/v1/auth/oauth/linked
```

## Troubleshooting

### Migration Issues
```bash
# Check migration status
make db-migrate-status

# Validate migration integrity
make db-migrate-validate

# If migrations are corrupted, reset (WARNING: loses data)
make db-migrate-reset-confirmed
```

### Air not working
```bash
# Reinstall Air
make install-tools

# Check Air configuration
cat .air.toml
```

### Database connection failed
```bash
# Check if PostgreSQL is running
make docker-up

# Verify connection details
make db-connect
```

### Port already in use
Change `SERVER_PORT` in your `.env` file or stop other processes using the port.

### Build errors
```bash
# Clean everything and start fresh
make clean-all
make deps
make build
```

## API Testing

### Using curl
```bash
# Check email availability
curl -X POST http://localhost:8888/api/v1/check-email \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Register user
curl -X POST http://localhost:8888/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8888/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Use token for protected endpoints
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/api/v1/conversations

# Send a message (creates new conversation)
curl -X POST http://localhost:8888/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello AI!","stream":false}'

# OAuth providers
curl http://localhost:8888/api/v1/auth/oauth/providers

# Linked accounts (requires auth)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/api/v1/auth/oauth/linked
```

### Health Check
```bash
curl http://localhost:8888/health
```

This should return `{"status":"healthy"}` if everything is working correctly.