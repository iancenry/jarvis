# Go Boilerplate Backend

A production-ready Go backend service built with Echo framework, featuring clean architecture, comprehensive middleware, and modern DevOps practices.

## Architecture Overview

This backend follows clean architecture principles with clear separation of concerns:

```
backend/
├── cmd/go-boilerplate/        # Application entry point
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── database/             # Database connections and migrations
│   ├── handler/              # HTTP request handlers
│   ├── service/              # Business logic layer
│   ├── repository/           # Data access layer
│   ├── model/                # Domain models
│   ├── middleware/           # HTTP middleware
│   ├── lib/                  # Shared libraries
│   └── validation/           # Request validation
├── static/                   # Static files (OpenAPI spec)
├── templates/                # Email templates
└── Taskfile.yml              # Task automation
```

## Features

### Core Framework
- **Echo v4**: High-performance, minimalist web framework
- **Clean Architecture**: Handlers → Services → Repositories → Models
- **Dependency Injection**: Constructor-based DI for testability

### Database
- **PostgreSQL**: Primary database with pgx/v5 driver
- **Migration System**: Tern for schema versioning
- **Connection Pooling**: Optimized for production workloads
- **Transaction Support**: ACID compliance for critical operations

### Authentication & Security
- **Clerk Integration**: Modern authentication service
- **JWT Validation**: Secure token verification
- **Role-Based Access**: Configurable permission system
- **Rate Limiting**: 20 requests/second per IP
- **Security Headers**: XSS, CSRF, and clickjacking protection

### Observability
- **New Relic APM**: Application performance monitoring
- **Structured Logging**: JSON logs with Zerolog
- **Request Tracing**: Distributed tracing support
- **Health Checks**: Readiness and liveness endpoints
- **Custom Metrics**: Business-specific monitoring

### Background Jobs
- **Asynq**: Redis-based distributed task queue
- **Priority Queues**: Critical, default, and low priority
- **Job Scheduling**: Cron-like task scheduling
- **Retry Logic**: Exponential backoff for failed jobs
- **Job Monitoring**: Real-time job status tracking

### Email Service
- **Resend Integration**: Reliable email delivery
- **HTML Templates**: Beautiful transactional emails
- **Preview Mode**: Test emails in development
- **Batch Sending**: Efficient bulk operations

### API Documentation
- **OpenAPI 3.0**: Complete API specification
- **Swagger UI**: Interactive API explorer
- **Auto-generation**: Code-first approach

## Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Task (taskfile.dev)

### Installation

1. Install dependencies:
```bash
go mod download
```

2. Set up environment:
```bash
cp .env.example .env
# Configure your environment variables
```

3. Run migrations:
```bash
task migrations:up
```

4. Start the server:
```bash
task run
```

## Configuration

Configuration is managed through environment variables with the `BOILERPLATE_` prefix:

### Database Configuration
```env
BOILERPLATE_DATABASE_HOST=localhost
BOILERPLATE_DATABASE_PORT=5432
BOILERPLATE_DATABASE_USER=postgres
BOILERPLATE_DATABASE_PASSWORD=password
BOILERPLATE_DATABASE_NAME=boilerplate
BOILERPLATE_DATABASE_SSL_MODE=disable
```

### Server Configuration
```env
BOILERPLATE_SERVER_PORT=8080
BOILERPLATE_SERVER_READ_TIMEOUT=10s
BOILERPLATE_SERVER_WRITE_TIMEOUT=10s
BOILERPLATE_SERVER_SHUTDOWN_TIMEOUT=30s
```

### Authentication
```env
BOILERPLATE_AUTH_CLERK_SECRET_KEY=sk_test_...
BOILERPLATE_AUTH_CLERK_PUBLISHABLE_KEY=pk_test_...
```

### Redis Configuration
```env
BOILERPLATE_REDIS_HOST=localhost
BOILERPLATE_REDIS_PORT=6379
BOILERPLATE_REDIS_PASSWORD=
BOILERPLATE_REDIS_DB=0
```

### Email Service
```env
BOILERPLATE_EMAIL_RESEND_API_KEY=re_...
BOILERPLATE_EMAIL_FROM_ADDRESS=noreply@example.com
BOILERPLATE_EMAIL_FROM_NAME=Go Boilerplate
```

### Observability
```env
BOILERPLATE_OBSERVABILITY_NEWRELIC_LICENSE_KEY=...
BOILERPLATE_OBSERVABILITY_NEWRELIC_APP_NAME=go-boilerplate
BOILERPLATE_OBSERVABILITY_LOG_LEVEL=info
```

## Development

### Available Tasks

```bash
task help                    # Show all available tasks
task run                     # Run the application
task test                    # Run tests
task test:integration        # Run integration tests
task migrations:new name=X   # Create new migration
task migrations:up           # Apply migrations
task migrations:down         # Rollback last migration
task tidy                    # Format and tidy dependencies
task lint                    # Run linters
task build                   # Build the application
```

### Project Structure

#### Handlers (`internal/handler/`)
HTTP request handlers that:
- Parse and validate requests
- Call appropriate services
- Format responses
- Handle HTTP-specific concerns

#### Services (`internal/service/`)
Business logic layer that:
- Implements use cases
- Orchestrates operations
- Enforces business rules
- Handles transactions

#### Repositories (`internal/repository/`)
Data access layer that:
- Encapsulates database queries
- Provides data mapping
- Handles database-specific logic
- Supports multiple data sources

#### Models (`internal/model/`)
Domain entities that:
- Define core business objects
- Include validation rules
- Remain database-agnostic

#### Middleware (`internal/middleware/`)
Cross-cutting concerns:
- Authentication/Authorization
- Request logging
- Error handling
- Rate limiting
- CORS
- Security headers

### Testing

#### Unit Tests
```bash
go test ./...
```

#### Integration Tests
```bash
# Requires Docker
go test -tags=integration ./...
```

#### Test Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### API Endpoints

#### Health Checks
- `GET /health` - Basic health check
- `GET /readiness` - Readiness probe
- `GET /liveness` - Liveness probe

#### Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Refresh token

#### User Management
- `GET /users` - List users
- `GET /users/:id` - Get user details
- `POST /users` - Create user
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user

#### API Documentation
- `GET /swagger/*` - Swagger UI
- `GET /openapi.json` - OpenAPI specification

## Error Handling

The application uses a structured error handling approach:

```go
// Custom error types
type ValidationError struct {
    Field   string
    Message string
}

type BusinessError struct {
    Code    string
    Message string
}

// Error responses
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input",
        "details": {
            "field": "email",
            "message": "Invalid email format"
        }
    }
}
```

## Logging

Structured logging with Zerolog:

```go
log.Info().
    Str("user_id", userID).
    Str("action", "login").
    Msg("User logged in successfully")
```

Log levels:
- `debug`: Detailed debugging information
- `info`: General informational messages
- `warn`: Warning messages
- `error`: Error messages
- `fatal`: Fatal errors that cause shutdown

## Deployment

### Docker

Build and run with Docker:

```bash
docker build -t go-boilerplate .
docker run -p 8080:8080 --env-file .env go-boilerplate
```

### Docker Compose

```bash
docker-compose up -d
```

### Production Checklist

- [ ] Set production environment variables
- [ ] Enable SSL/TLS
- [ ] Configure production database
- [ ] Set up monitoring alerts
- [ ] Configure log aggregation
- [ ] Enable rate limiting
- [ ] Set up backup strategy
- [ ] Configure auto-scaling
- [ ] Implement graceful shutdown
- [ ] Set up CI/CD pipeline

## Performance Optimization

### Database
- Connection pooling configured
- Prepared statements for frequent queries
- Indexes on commonly queried fields
- Query optimization with EXPLAIN ANALYZE

### Caching
- Redis for session storage
- In-memory caching for hot data
- HTTP caching headers

### Concurrency
- Goroutine pools for parallel processing
- Context-based cancellation
- Proper mutex usage

## Security Best Practices

1. **Input Validation**: All inputs validated and sanitized
2. **SQL Injection**: Parameterized queries only
3. **XSS Protection**: Output encoding and CSP headers
4. **CSRF Protection**: Token-based protection
5. **Rate Limiting**: Per-IP and per-user limits
6. **Secrets Management**: Environment variables, never in code
7. **HTTPS Only**: Enforce TLS in production
8. **Dependency Scanning**: Regular vulnerability checks

## Troubleshooting

### Common Issues

#### Database Connection Errors
```bash
# Check PostgreSQL is running
docker-compose ps

# Verify connection string
psql $BOILERPLATE_DATABASE_URL
```

#### Migration Failures
```bash
# Check migration status
task migrations:status

# Rollback if needed
task migrations:down
```

#### Performance Issues
```bash
# Enable debug logging
BOILERPLATE_OBSERVABILITY_LOG_LEVEL=debug task run

# Check New Relic dashboard
# Review slow query logs
```

## Contributing

1. Follow Go best practices and idioms
2. Write tests for new features
3. Update documentation
4. Run linters before committing
5. Keep commits atomic and well-described

## License

See the parent project's LICENSE file.