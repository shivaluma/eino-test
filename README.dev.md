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
- `make db-migrate` - Run database migrations
- `make db-reset` - Reset database and run migrations
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
# Apply migrations
make db-migrate

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

## Troubleshooting

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
```

### Health Check
```bash
curl http://localhost:8888/health
```

This should return `{"status":"healthy"}` if everything is working correctly.