# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a full-stack web application built with **Go backend** and **Next.js frontend** that provides secure OAuth authentication and AI-powered chat functionality. The application demonstrates modern web development practices with:

- **Secure OAuth 2.0 Flow**: Multi-provider authentication (GitHub, Google) with PKCE support
- **JWT Token Management**: Access and refresh tokens with HTTP-only cookies
- **AI Integration**: Chat functionality using the Eino AI framework with OpenAI
- **Modern Frontend**: Next.js 15 with React 19, TypeScript, and Tailwind CSS
- **Database**: PostgreSQL with proper migrations and relationship management
- **Security**: Comprehensive security measures following OWASP guidelines

## Technology Stack

### Backend (Go)
- **Language**: Go 1.24.5
- **Web Framework**: Echo v4 (Labstack)
- **AI Framework**: CloudWego Eino v0.4.0 with OpenAI extension
- **Database**: PostgreSQL with pgx/v5 driver
- **Authentication**: JWT tokens with golang.org/x/oauth2
- **Validation**: go-playground/validator/v10
- **Security**: bcrypt for password hashing, UUID for IDs
- **Configuration**: godotenv for environment management

### Frontend (Next.js)
- **Framework**: Next.js 15 with App Router and Turbopack
- **React**: v19.1.1 with Server Components
- **Language**: TypeScript 5.9.2 with strict mode
- **Styling**: Tailwind CSS v4.1.11
- **UI Components**: Radix UI with shadcn/ui design system
- **State Management**: TanStack React Query v5.84.1
- **Internationalization**: next-intl v4.3.4
- **Animations**: Framer Motion v12.23.12
- **Icons**: Lucide React
- **Code Quality**: Biome v1.9.4 (linting + formatting)

### Database & Infrastructure
- **Database**: PostgreSQL with UUID primary keys
- **Migrations**: Custom Go migration system
- **ORM**: Raw SQL with pgx for performance
- **Caching**: In-memory session storage
- **Logging**: Structured logging with context
- **Containerization**: Docker with multi-stage builds

### Development Tools
- **Package Managers**: Go modules, pnpm for frontend
- **Build Tools**: Go build, Next.js build system
- **Code Quality**: go vet, go fmt, Biome (frontend)
- **Testing**: Go test, planned frontend testing with Vitest
- **Hot Reload**: Air for Go, Next.js dev server

## Project Architecture

### Directory Structure
```
eino-test/
├── cmd/                          # Application entry points
│   ├── server/main.go           # Main server application
│   └── migrate/main.go          # Database migration tool
├── config/                       # Configuration management
│   └── config.go                # Environment-based config
├── internal/                     # Private application code
│   ├── auth/                    # Authentication services
│   │   ├── auth.go             # JWT service
│   │   └── oauth.go            # OAuth providers
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth_handler.go     # Authentication endpoints
│   │   ├── oauth_handler.go    # OAuth flow handlers
│   │   └── conversation_handler.go # Chat endpoints
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go             # JWT validation
│   │   └── logging.go          # Request logging
│   ├── models/                  # Data models
│   │   ├── user.go             # User and OAuth models
│   │   └── conversation.go     # Chat models
│   ├── repository/              # Data access layer
│   │   ├── user_repository.go  # User data operations
│   │   ├── oauth_repository.go # OAuth data operations
│   │   └── conversation_repository.go # Chat data
│   ├── database/                # Database connection
│   ├── logger/                  # Logging utilities
│   └── ai/                      # AI integration
│       ├── service.go          # AI service interface
│       ├── providers/          # AI provider implementations
│       └── templates/          # Prompt templates
├── frontend/                     # Next.js application
│   ├── app/                     # App Router structure
│   │   ├── (auth)/             # Authentication pages
│   │   ├── (chat)/             # Main application
│   │   └── api/                # API routes
│   ├── components/              # React components
│   │   ├── ui/                 # Reusable UI components
│   │   ├── layouts/            # Layout components
│   │   └── icons/              # Icon components
│   ├── lib/                     # Frontend utilities
│   │   ├── auth/               # Auth utilities
│   │   ├── api/                # API clients
│   │   └── utils.ts            # Common utilities
│   ├── types/                   # TypeScript types
│   ├── hooks/                   # Custom React hooks
│   └── middleware.ts            # Next.js middleware
├── migrations/                   # Database schema migrations
├── scripts/                      # Build and utility scripts
└── docker-compose.yml           # Development environment
```

### Backend Architecture Patterns

#### Layered Architecture
```
┌─────────────────────────────────────────────┐
│                HTTP Layer                   │
│  (handlers, middleware, routing)            │
├─────────────────────────────────────────────┤
│              Business Logic                 │
│     (auth service, ai service)              │
├─────────────────────────────────────────────┤
│               Data Layer                    │
│        (repositories, models)               │
├─────────────────────────────────────────────┤
│              Infrastructure                 │
│     (database, logging, config)             │
└─────────────────────────────────────────────┘
```

#### Dependency Injection Pattern
- Services are injected into handlers
- Repositories are injected into services
- Database connections are injected into repositories
- Configuration is injected at startup

#### Repository Pattern
- Abstract data access behind interfaces
- Separate concerns between business logic and data storage
- Enable easier testing with mock repositories
- Support for different database backends

### Frontend Architecture

#### Next.js App Router Structure
- **Route Groups**: `(auth)` and `(chat)` for logical organization
- **Server Components**: Default for better performance
- **Client Components**: Used only when interactivity is needed
- **API Routes**: Backend integration points
- **Middleware**: Route protection and authentication

#### Component Architecture
```
┌─────────────────────────────────────────────┐
│                 Pages                       │
│           (Route handlers)                  │
├─────────────────────────────────────────────┤
│              Layouts                        │
│        (Shared UI structure)                │
├─────────────────────────────────────────────┤
│           Feature Components                │
│      (Auth forms, Chat interface)           │
├─────────────────────────────────────────────┤
│             UI Components                   │
│         (Reusable elements)                 │
└─────────────────────────────────────────────┘
```

## Code Style Guidelines

### Backend (Go) Style

#### General Conventions
- **Formatting**: Use `go fmt` for consistent formatting
- **Naming**: Follow Go naming conventions (camelCase for private, PascalCase for public)
- **Line Length**: Keep lines under 80-100 characters when practical
- **Imports**: Group standard library, third-party, and local imports separately
- **Comments**: Use godoc-style comments for public functions and types

#### Package Organization
```go
// Package structure example
package handlers

import (
    // Standard library
    "context"
    "net/http"
    
    // Third-party packages
    "github.com/labstack/echo/v4"
    
    // Local packages
    "github.com/shivaluma/eino-agent/internal/auth"
    "github.com/shivaluma/eino-agent/internal/models"
)
```

#### Error Handling
```go
// Always handle errors explicitly
user, err := h.userRepo.GetByID(ctx, userID)
if err != nil {
    log.Error().Err(err).Msg("Failed to get user")
    return c.JSON(http.StatusInternalServerError, map[string]string{
        "error": "User not found",
    })
}
```

#### Struct Definitions
```go
// Use struct tags for JSON serialization and validation
type User struct {
    ID       uuid.UUID `json:"id" db:"id"`
    Username string    `json:"username" db:"username" validate:"required,min=3,max=50"`
    Email    string    `json:"email" db:"email" validate:"required,email"`
    Password string    `json:"-" db:"password_hash"` // Never serialize passwords
}
```

#### Database Queries
```go
// Use prepared statements and context
const query = `
    SELECT id, username, email, created_at 
    FROM users 
    WHERE email = $1`

var user models.User
err := r.db.QueryRowContext(ctx, query, email).Scan(
    &user.ID, &user.Username, &user.Email, &user.CreatedAt,
)
```

### Frontend (TypeScript/React) Style

#### General Conventions
- **Formatting**: 2-space indentation, semicolons, single quotes
- **Naming**: camelCase for variables/functions, PascalCase for components
- **File Extensions**: `.tsx` for React components, `.ts` for utilities
- **Path Aliases**: Use `@/` for src root imports

#### Component Structure
```typescript
// Component file structure
'use client'; // Only when client interactivity is needed

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import type { User } from '@/types/authentication';

interface SignInFormProps {
  onSubmit: (credentials: LoginCredentials) => void;
  loading?: boolean;
}

export function SignInForm({ onSubmit, loading = false }: SignInFormProps) {
  // Component logic here
  return (
    // JSX here
  );
}
```

#### API Client Pattern
```typescript
// Consistent API error handling
export const apiClient = {
  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        credentials: 'include', // Include cookies
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      
      return { data: await response.json(), error: null };
    } catch (error) {
      return { data: null, error: error.message };
    }
  },
};
```

#### Type Definitions
```typescript
// Shared types for frontend/backend communication
export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  created_at: string;
}

export interface OAuthProvider {
  name: 'github' | 'google';
  enabled: boolean;
  client_id: string;
}
```

### Database Schema Conventions

#### Table Design
- **Primary Keys**: Always use UUID v4 for better distribution
- **Timestamps**: Include `created_at` and `updated_at` on all tables
- **Foreign Keys**: Use proper CASCADE settings
- **Indexing**: Index frequently queried columns

```sql
-- Example table structure
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

#### Migration Patterns
- **Sequential**: Number migrations sequentially (001_, 002_, etc.)
- **Descriptive**: Include timestamp and description in filename
- **Reversible**: Include DOWN migrations when possible
- **Data Safety**: Never drop columns without proper backup

## Development Commands

### Backend Commands
```bash
# Development
go run cmd/server/main.go              # Start development server
go run cmd/migrate/main.go up          # Run database migrations
go run cmd/migrate/main.go down        # Rollback migrations

# Building
go build -o server cmd/server/main.go  # Build server binary
go build -o migrate cmd/migrate/main.go # Build migration tool

# Testing & Quality
go test ./...                          # Run all tests
go test -v ./internal/handlers         # Run specific package tests
go test -race ./...                    # Test with race detection
go vet ./...                          # Static analysis
go fmt ./...                          # Format code
golint ./...                          # Linting (if installed)

# Dependencies
go mod tidy                           # Clean up dependencies
go mod download                       # Download dependencies
go mod vendor                         # Create vendor directory
```

### Frontend Commands
```bash
# Development
cd frontend
pnpm dev                              # Start dev server with Turbopack
pnpm build                            # Production build
pnpm start                            # Start production server

# Code Quality
pnpm lint                             # Run ESLint and Biome
pnpm lint:fix                         # Fix linting issues
pnpm format                           # Format with Biome
pnpm check-types                      # TypeScript type checking

# Testing
pnpm test                             # Run tests (when implemented)
pnpm test:watch                       # Watch mode testing

# Dependencies
pnpm install                          # Install dependencies
pnpm update                           # Update dependencies
pnpm outdated                         # Check outdated packages
```

### Docker Commands
```bash
# Development environment
docker-compose up -d                  # Start all services
docker-compose down                   # Stop all services
docker-compose logs -f api            # Follow API logs

# Database only
docker run --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 -d postgres:15

# Build production images
docker build -t eino-backend .
docker build -t eino-frontend ./frontend
```

## Security Considerations

### Authentication & Authorization
- **JWT Tokens**: Short-lived access tokens (15 minutes), longer refresh tokens (7 days)
- **HTTP-only Cookies**: Tokens stored in secure, HTTP-only cookies to prevent XSS
- **CSRF Protection**: SameSite cookie policy and state parameter validation
- **OAuth Security**: PKCE support for enhanced security, state parameter validation

### Input Validation
- **Server-side Validation**: All inputs validated using go-playground/validator
- **SQL Injection Prevention**: Parameterized queries with pgx
- **XSS Prevention**: Proper output encoding and CSP headers
- **Data Sanitization**: User inputs sanitized before database storage

### Infrastructure Security
- **Database Security**: Connection pooling, prepared statements, proper indexing
- **CORS Configuration**: Restrictive CORS policy for production
- **Logging**: Security events logged with context
- **Error Handling**: Generic error messages to prevent information disclosure

## Testing Strategy

### Backend Testing
```go
// Example test structure
func TestOAuthHandler(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Create test dependencies
    userRepo := repository.NewUserRepository(db)
    oauthRepo := repository.NewOAuthRepository(db)
    
    // Test OAuth flow
    t.Run("InitiateOAuth", func(t *testing.T) {
        // Test implementation
    })
}
```

### Frontend Testing (Planned)
- **Unit Tests**: Component testing with React Testing Library
- **Integration Tests**: API integration testing
- **E2E Tests**: End-to-end user flows with Playwright

## Deployment

### Local Development
```bash
# Start database
docker-compose up -d postgres

# Run migrations
go run cmd/migrate/main.go up

# Start backend
go run cmd/server/main.go

# Start frontend (in another terminal)
cd frontend && pnpm dev
```

### Production Deployment
```bash
# Build backend
go build -o server cmd/server/main.go

# Build frontend
cd frontend && pnpm build

# Run with production environment
ENV=production ./server
```

### Docker Deployment
```bash
# Full stack deployment
docker-compose up -d

# Individual services
docker-compose up -d postgres  # Database only
docker-compose up -d api       # Backend only
docker-compose up -d frontend  # Frontend only
```

## Key Dependencies & Libraries

### Backend Dependencies
- **Echo v4**: High-performance HTTP web framework
- **pgx/v5**: Fast, feature-rich PostgreSQL driver
- **golang.org/x/oauth2**: OAuth 2.0 client library
- **lestrrat-go/jwx/v2**: JWT library with comprehensive features
- **google/uuid**: UUID generation
- **go-playground/validator**: Struct validation
- **godotenv**: Environment variable loading
- **zerolog**: Structured, fast logging

### Frontend Dependencies
- **Next.js 15**: React framework with App Router
- **React 19**: Latest React with concurrent features
- **TypeScript**: Static type checking
- **Tailwind CSS**: Utility-first CSS framework
- **Radix UI**: Accessible component primitives
- **TanStack Query**: Server state management
- **Framer Motion**: Animation library
- **next-intl**: Internationalization
- **Biome**: Fast linter and formatter

## Best Practices

### Code Organization
- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Injection**: Explicit dependencies for testability
- **Interface Usage**: Abstract implementations behind interfaces
- **Error Handling**: Consistent error handling patterns
- **Logging**: Structured logging with context

### Performance
- **Database Indexing**: Proper indexes on frequently queried columns
- **Connection Pooling**: Efficient database connection management
- **Caching**: Strategic caching of frequently accessed data
- **Static Assets**: CDN and caching for frontend assets

### Maintainability
- **Documentation**: Comprehensive code comments and documentation
- **Testing**: Unit and integration tests for critical paths
- **Code Review**: Peer review process for all changes
- **Monitoring**: Application and infrastructure monitoring

This application demonstrates modern full-stack development practices with a focus on security, performance, and maintainability. The architecture supports easy scaling and extension with additional features.