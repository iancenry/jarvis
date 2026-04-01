# Go JARVIS Backend

A production-ready Go backend service built with Echo framework, featuring clean architecture, comprehensive middleware, and modern DevOps practices.

## Architecture Overview

This backend follows clean architecture principles with clear separation of concerns:

```
backend/
├── cmd/go-jarvis/             # API entry point
├── cmd/worker/                # Background worker entry point
├── cmd/cron/                  # One-shot cron scheduler entry point
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── database/             # Database connections and migrations
│   ├── handler/              # HTTP request handlers
│   ├── service/              # Business logic layer
│   ├── repository/           # Data access layer
│   ├── model/                # Domain models
│   ├── middleware/           # HTTP middleware
│   ├── lib/                  # Shared libraries - reusable, domain-agnostic infrastructure code
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

- Go 1.24+
- PostgreSQL 16+
- Redis 8+
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

4. Start the API server:

```bash
task run
```

If you need background jobs locally, start the worker in a second terminal:

```bash
task worker
```

## Configuration

Configuration is managed through environment variables with the `JARVIS_` prefix:

## Development

### Available Tasks

```bash
task help                    # Show all available tasks
task run                     # Run the API server
task worker                  # Run the background worker
task dev:all                 # Run the API server and worker together
task cron:list               # List available cron jobs
task cron:run job=JOB_NAME   # Run one cron job manually
task test                    # Run tests
task migrations:new name=X   # Create new migration
task migrations:up           # Apply migrations
task tidy                    # Format and tidy dependencies
```

### Local Runtime Modes

`task run` starts only the HTTP API. This is the right default when you are working on request/response behavior and do not need queued jobs.

`task worker` starts the dedicated Asynq consumer. Run it alongside `task run` when you want reminder emails, reports, or other queued work to be processed locally.

`task dev:all` starts both the API and worker in one terminal session. It does not run cron jobs automatically.

Cron remains a one-shot command. Use `task cron:list` to see the available job names, then run one manually, for example:

```bash
task cron:run job=due-date-reminders
task cron:run job=overdue-notifications
task cron:run job=weekly-reports
task cron:run job=auto-archive
```

For local background flows:

- API requires PostgreSQL.
- Worker requires PostgreSQL and Redis.
- Cron commands require PostgreSQL and Redis.

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

## Contributing

1. Follow Go best practices and idioms
2. Write tests for new features
3. Update documentation
4. Run linters before committing
5. Keep commits atomic and well-described

## License

See the parent project's LICENSE file.
