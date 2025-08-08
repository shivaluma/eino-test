# Deployment Guide

This guide covers deploying the AI Food Agent application with proper database migration handling.

## Docker Deployment

### Single Container (Recommended)
The Dockerfile has been updated to include the migrations directory:

```bash
# Build the image
docker build -t food-agent:latest .

# Run with environment variables
docker run -d \
  --name food-agent \
  -p 8888:8888 \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=your-db-name \
  -e JWT_ACCESS_SECRET=your-access-secret \
  -e JWT_REFRESH_SECRET=your-refresh-secret \
  food-agent:latest
```

### Docker Compose
```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f food-agent-api
```

## Migration Handling in Production

### Automatic Migrations (Default)
The application automatically runs migrations on startup. This is enabled by default in `cmd/server/main.go`:

```go
// Run database migrations on startup
log.Println("Running database migrations...")
migrator := migrations.NewMigrator(db.Pool, "migrations", cfg)
if err := migrator.Migrate(ctx); err != nil {
    log.Fatalf("Failed to run database migrations: %v", err)
}
```

### Manual Migration Control
To disable automatic migrations and run them manually:

1. **Disable auto-migration** in your deployment:
   ```go
   // Comment out or remove the auto-migration code in cmd/server/main.go
   ```

2. **Run migrations manually**:
   ```bash
   # Build migration CLI
   docker run --rm -v $(pwd):/app -w /app golang:1.24.5-alpine \
     go build -o migrate cmd/migrate/main.go
   
   # Run migrations
   ./migrate -command=migrate
   ```

## Kubernetes Deployment

### Using Init Containers
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: food-agent
spec:
  template:
    spec:
      initContainers:
      - name: migrations
        image: food-agent:latest
        command: ['/root/server']
        env:
        - name: RUN_MIGRATIONS_ONLY
          value: "true"
        # Add your database env vars here
      containers:
      - name: food-agent
        image: food-agent:latest
        # Add your config here
```

### ConfigMap for Migrations
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: migrations
data:
  # Include your migration files here
  000_migration_system.sql: |
    -- Migration content here
  001_20250108000001_initial_schema.sql: |
    -- Migration content here
```

## Cloud Deployment

### Heroku
1. **Create Procfile**:
   ```
   release: ./migrate -command=migrate
   web: ./server
   ```

2. **Deploy**:
   ```bash
   git push heroku main
   ```

### Railway/Render
Add migration command to your service configuration:
```json
{
  "buildCommand": "go build -o server cmd/server/main.go",
  "startCommand": "./migrate -command=migrate && ./server"
}
```

### AWS/GCP/Azure
Use the Docker image with environment variables for database connection.

## Database Setup

### PostgreSQL Requirements
- PostgreSQL 12+ (recommended 15+)
- Required extensions: `uuid-ossp` (automatically created by migrations)
- User with CREATE, DROP, and SELECT permissions on `public` schema

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database
DB_SSL_MODE=require  # Use 'disable' for local development

# JWT
JWT_ACCESS_SECRET=your-very-secure-access-secret-min-32-chars
JWT_REFRESH_SECRET=your-very-secure-refresh-secret-min-32-chars
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h

# Server
SERVER_PORT=8888
SERVER_HOST=0.0.0.0

# OAuth (optional)
OAUTH_GITHUB_CLIENT_ID=your_github_client_id
OAUTH_GITHUB_CLIENT_SECRET=your_github_client_secret
OAUTH_GOOGLE_CLIENT_ID=your_google_client_id
OAUTH_GOOGLE_CLIENT_SECRET=your_google_client_secret
OAUTH_STATE_SECRET=your-oauth-state-secret
OAUTH_FRONTEND_URL=https://your-frontend-domain.com

# AI (optional)
OPENAI_API_KEY=your_openai_api_key
OPENAI_MODEL_NAME=gpt-3.5-turbo
OPENAI_BASE_URL=https://api.openai.com/v1
```

## Migration Troubleshooting

### Common Issues

1. **"migrations directory not found"**
   - Ensure Dockerfile copies migrations: `COPY --from=builder /app/migrations ./migrations`
   - Verify working directory in container matches migration path

2. **Migration checksum errors**
   - Don't modify applied migration files
   - Use `make db-migrate-validate` to check integrity
   - Reset if necessary: `make db-migrate-reset-confirmed`

3. **Database connection failed**
   - Check database credentials and network connectivity
   - Verify SSL mode settings
   - Ensure database exists and user has permissions

4. **Migration stuck/partial failure**
   - Check migration logs for specific error
   - Use `make db-migrate-status` to see current state
   - Manual fix may be required for failed migrations

### Migration Status Check
```bash
# Check current migration status
docker exec -it your-container /root/migrate -command=status

# Validate migration integrity
docker exec -it your-container /root/migrate -command=validate
```

### Emergency Rollback
```bash
# Rollback last migration
docker exec -it your-container /root/migrate -command=rollback

# Rollback to specific version
docker exec -it your-container /root/migrate -command=rollback-to -version=2
```

## Security Considerations

### Production Settings
- Use strong JWT secrets (min 32 characters)
- Enable SSL/TLS for database connections
- Use environment variables, not hardcoded secrets
- Enable CORS only for trusted domains
- Use secure OAuth redirect URLs (HTTPS)

### Database Security
- Use database user with minimal required permissions
- Enable PostgreSQL SSL
- Regular database backups
- Monitor for unusual migration activity

## Monitoring

### Health Checks
```bash
# Application health
curl http://your-domain/health

# Database health (included in above)
```

### Migration Monitoring
- Monitor startup logs for migration success/failure
- Set up alerts for migration failures
- Track migration execution time
- Regular validation of migration checksums

## Backup Strategy

### Before Deployments
```bash
# Create database backup before migration
pg_dump your_database > backup_$(date +%Y%m%d_%H%M%S).sql

# Test migration on backup first
```

### Regular Backups
- Automated daily database backups
- Store migration history and checksums
- Test restore procedures regularly