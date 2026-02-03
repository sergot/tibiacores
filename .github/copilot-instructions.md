---
description: TibiaCores project context and architectural patterns for AI coding assistants.
---
# TibiaCores Project Context

## Project Identity
- **Name**: TibiaCores
- **Purpose**: Web application for tracking and managing soul cores in the MMORPG Tibia
- **Stack**: Go (Backend), Vue 3 + TypeScript (Frontend), PostgreSQL (Database)

## Quick Reference

For detailed coding standards, contribution workflows, and architectural patterns, see:
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Complete coding standards, patterns, and workflows
- **[docs/database.md](../docs/database.md)** - Database schema and migration procedures
- **[docs/setup.md](../docs/setup.md)** - Development environment setup

## Critical Rules

### Error Handling
**ALWAYS use `apperror` package** - Never return raw errors:
```go
return apperror.DatabaseError("failed to create user", err).
    WithDetails(&apperror.DatabaseErrorDetails{
        Operation: "CreateUser",
        Table:     "users",
    })
```

### Logging
**ALWAYS use `log/slog`** for structured logging:
```go
slog.Info("user created", "user_id", userID)
slog.Error("database error", "error", err, "operation", "CreateUser")
```

### Context
**ALWAYS pass context** to database operations:
```go
user, err := store.GetUser(ctx, userID)  // ✓ Correct
```

### Database Changes
1. Create migration: `goose -dir db/migrations create feature_name sql`
2. Write SQL in `db/queries/*.sql`
3. Generate code: `sqlc generate`
4. Use in handlers

### External HTTP Clients
**MUST configure 10s timeout and transport**:
```go
client := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}
```

### External Services
- **Mailgun**: Hardcoded to EU region (`mailgun.APIBaseEU`) - credentials must match
- **TibiaData**: Rate limits apply, handle gracefully
- **EmailOctopus**: Newsletter subscriptions

### Background Jobs
- Current implementation: Single-instance claim checker (15-min interval)
- **TODO**: Add Postgres advisory locks for horizontal scaling
- Must include panic recovery and stack trace logging

## Project Layout
```
backend/
├── cmd/server/       # Entry point, routing
├── handlers/         # HTTP handlers (controllers)
├── services/         # External integrations
├── auth/             # JWT, OAuth, passwords
├── db/
│   ├── migrations/   # Goose migrations
│   ├── queries/      # SQL (sqlc source)
│   └── sqlc/         # Generated Go code
├── middleware/       # Logging, errors
└── pkg/
    ├── apperror/     # Custom error system
    └── validator/    # Request validation

frontend/
├── src/
│   ├── components/   # Vue components
│   ├── views/        # Pages
│   ├── stores/       # Pinia state
│   ├── services/     # API clients
│   └── router/       # Routes
```

## Common Tasks

### Adding an Endpoint
1. Define SQL in `db/queries/*.sql`
2. Run `sqlc generate`
3. Create handler in `handlers/`
4. Register route in `cmd/server/main.go` (setupRoutes function)

### Testing
- Backend: `go test ./...`
- Use gomock for store mocking: `db/mock/store.go`
- Frontend: `npm run test`

For comprehensive guidance, always refer to [CONTRIBUTING.md](../CONTRIBUTING.md).
