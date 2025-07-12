# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wastebin is a self-hosted web service for sharing pastes anonymously. It consists of:
- **Backend**: Go application using Chi router, GORM ORM, and Zap logging
- **Frontend**: React SPA with TypeScript, Material-UI, and Vite (recently migrated from SvelteKit)
- **Database**: PostgreSQL (production) or SQLite (development)

## Common Development Commands

### Backend Development
```bash
# Run the backend in development mode (uses SQLite)
WASTEBIN_DEV=true go run cmd/wastebin/main.go

# Run all tests with coverage
go test -v -race -covermode atomic -coverprofile=covprofile ./...

# Run specific package tests
go test -v ./handlers/
go test -v ./storage/
go test -v ./config/

# Build the binary
go build -o wastebin cmd/wastebin/main.go
```

### Frontend Development
```bash
# Navigate to web directory first
cd web

# Install dependencies
npm install

# Run development server (hot reload on http://localhost:5173)
npm run dev

# Build for production
npm run build

# Run linter
npm run lint

# Preview production build
npm run preview
```

### Docker Operations
```bash
# Build multi-platform Docker image
./build.sh

# Run with Docker Compose (includes PostgreSQL)
docker-compose up
```

## Architecture and Code Structure

### Backend Architecture

1. **Entry Point**: `cmd/wastebin/main.go` initializes config, database, routes, and starts the server with graceful shutdown handling.

2. **Route Registration**: `routes/routes.go` sets up Chi router with middleware stack:
   - Request ID generation
   - Structured JSON logging with Zap
   - Panic recovery
   - CORS handling
   - Static file serving for frontend

3. **API Endpoints** (all under `/api/v1/`):
   - `POST /api/v1/paste` - Create paste (form data: content, language, expiryTime, burn)
   - `GET /api/v1/paste/{uuid}` - Get paste as JSON
   - `DELETE /api/v1/paste/{uuid}` - Delete paste
   - `GET /paste/{uuid}/raw` - Get raw paste content (outside API namespace)

4. **Request Flow**:
   - Handlers in `handlers/paste.go` process requests
   - Storage operations through `storage/gorm.go` 
   - Models defined in `models/models.go`
   - Configuration loaded from environment variables via `config/config.go`

5. **Database Operations**:
   - GORM ORM with automatic migrations
   - Support for PostgreSQL and SQLite
   - Connection pooling configured
   - Paste model includes: UUID, Content, Burn flag, Language, ExpiryTimestamp

### Frontend Architecture

1. **Tech Stack**: React 18 with TypeScript, Material-UI components, React Router, Vite bundler

2. **Key Components**:
   - `src/App.tsx` - Main application component with routing
   - `src/pages/Home.tsx` - Paste creation form
   - `src/pages/Paste.tsx` - Paste viewing with syntax highlighting
   - `src/components/` - Reusable UI components

3. **Build Output**: Frontend builds to `web/dist/` which is embedded in the Go binary for production

## Testing Approach

### Backend Testing
- **Framework**: Go's built-in testing with testify assertions and comprehensive test utilities
- **Test Types**:
  - Unit tests for all handlers, config, storage, and utility functions
  - Integration tests for full API workflows and database operations
  - Security tests for XSS, SQL injection, rate limiting, and input validation
  - Performance tests and benchmarks for load testing and optimization
- **Test Database**: In-memory SQLite for unit tests, PostgreSQL for integration tests
- **Test Utilities**: Custom test server with helper functions in `internal/testutil/`

### Test Commands
```bash
# Run all tests with coverage
make test-coverage

# Run specific test types
make test                    # Unit tests only
make test-integration       # Integration tests with PostgreSQL
make test-security          # Security and vulnerability tests  
make test-performance       # Load and performance tests
make benchmark              # Performance benchmarks

# Test with coverage threshold (75%)
./scripts/test-coverage.sh

# Full test suite
make test-all
```

### Frontend Testing
- **Framework**: Jest and React Testing Library (configured for React migration)
- **Commands**: `npm run test` (in web directory)
- **Coverage**: Included in CI/CD pipeline

## Configuration

All configuration via environment variables with `WASTEBIN_` prefix:
- `WASTEBIN_WEBAPP_PORT` - Server port (default: 3000)
- `WASTEBIN_DB_HOST/PORT/USER/PASSWORD/NAME` - PostgreSQL connection
- `WASTEBIN_DEV` - Development mode (uses SQLite)
- `WASTEBIN_LOG_LEVEL` - Logging verbosity

## Recent Improvements (July 2025)

The codebase has been significantly improved with the following enhancements:

### 1. **Error Handling & API Responses**
- Structured error responses with consistent JSON format
- Custom error types for better error categorization
- Proper HTTP status codes and detailed error messages

### 2. **Security Enhancements** 
- Rate limiting (100 requests/minute per IP)
- Security headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection)
- Input validation with size limits (10MB max paste size)
- Improved CORS configuration with environment-based origins

### 3. **Database Improvements**
- Connection retry logic with exponential backoff
- Database health checks with timeout handling
- Improved connection pooling configuration
- Graceful connection closing with timeout

### 4. **Context Support**
- All database operations now use context for timeout handling
- Request context propagation throughout the application
- Proper cancellation support

### 5. **Configuration Validation**
- Comprehensive config validation on startup
- Required field checking based on environment
- Port and connection pool validation

### 6. **Dependency Injection**
- New `server` package with proper dependency management
- Eliminated global variables where possible
- Clean separation of concerns

### 7. **Enhanced Logging**
- Structured JSON logging with request IDs
- Request/response logging with full traceability
- Different log levels based on HTTP status codes
- User agent and referer tracking

### 8. **Graceful Shutdown**
- Improved shutdown handling with 30-second timeout
- Proper cleanup of database connections
- Signal handling for SIGINT and SIGTERM

### 9. **API Versioning**
- Support for API versioning via headers
- Content negotiation support
- Version information in response headers

### 10. **Health Monitoring**
- `/health` endpoint for basic service health
- `/health/db` endpoint for database health checks
- Comprehensive health check with database query testing

## CI/CD Pipeline

### GitHub Actions Workflow
The project uses a comprehensive CI/CD pipeline with the following jobs:

1. **Lint**: golangci-lint with custom configuration
2. **Security**: Gosec security scanning and vulnerability checks
3. **Unit Tests**: Multi-version Go testing (1.22, 1.23) with coverage reporting
4. **Integration Tests**: Full API testing with PostgreSQL
5. **Frontend Tests**: Node.js build and lint checking
6. **Build**: Multi-platform Docker builds (amd64, arm64)
7. **Performance Tests**: Automated benchmarking
8. **Dependency Check**: govulncheck for known vulnerabilities
9. **Release**: GoReleaser for tagged releases

### Quality Gates
- **Test Coverage**: Minimum 75% required
- **Security**: No high/critical vulnerabilities allowed
- **Performance**: Response time benchmarks tracked
- **Code Quality**: golangci-lint must pass
- **Build**: Multi-platform Docker builds must succeed

### Development Workflow
```bash
# Setup development environment
make setup

# Run tests locally
make check                  # Lint, security, tests
make test-coverage         # Full test suite with coverage

# Development server
make dev                   # Backend with hot reload
make frontend-dev          # Frontend with hot reload

# Build and deploy
make build-all            # Build frontend and backend
make docker-build         # Create Docker image
```

## Current Migration Status

The project has completed migration from Fiber to Chi router framework. The frontend migration from SvelteKit to React is ongoing. All backend improvements and comprehensive testing are now in place and the application is production-ready with enterprise-grade CI/CD pipeline.