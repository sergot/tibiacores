# Contributing to TibiaCores

Thank you for your interest in contributing to TibiaCores! This document provides comprehensive guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## How to Contribute

There are several ways you can contribute to TibiaCores:

1. **Report bugs**: Submit issues for any bugs you encounter
2. **Suggest features**: Propose new features or improvements  
3. **Submit pull requests**: Implement new features or fix bugs
4. **Improve documentation**: Help enhance the documentation
5. **Improve translations**: Contribute to the localization of the application
6. **Code review**: Review pull requests from other contributors

## Development Setup

### Quick Start

1. **Fork and clone** the repository
2. **Follow the setup guide**: See [docs/setup.md](docs/setup.md) for complete instructions including:
   - Environment configuration
   - Docker Compose setup
   - OAuth provider registration
   - Email service configuration
   - Database migrations
3. **Create a feature branch**: `git checkout -b feature/your-feature-name`

**Prerequisites**: Go 1.25+, Node.js 20+, Docker & Docker Compose

## Project Structure

\`\`\`
tibiacores/
├── backend/              # Go backend
│   ├── cmd/server/       # Application entry point
│   ├── handlers/         # HTTP handlers (controllers)
│   ├── services/         # Business logic & external integrations
│   ├── auth/             # Authentication (JWT, OAuth, passwords)
│   ├── db/
│   │   ├── migrations/   # Database schema migrations (Goose)
│   │   ├── queries/      # SQL queries (sqlc source)
│   │   └── sqlc/         # Generated Go code from sqlc
│   ├── middleware/       # HTTP middleware (logging, errors)
│   └── pkg/
│       ├── apperror/     # Custom error handling system
│       └── validator/    # Request validation
├── frontend/             # Vue 3 + TypeScript frontend
│   ├── src/
│   │   ├── components/   # Reusable Vue components
│   │   ├── views/        # Page-level components
│   │   ├── stores/       # Pinia state management
│   │   ├── services/     # API client services
│   │   ├── router/       # Vue Router configuration
│   │   └── i18n/         # Internationalization
└── docs/                 # Project documentation
\`\`\`

## Coding Standards

### Backend (Go 1.25+)

#### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use \`gofmt\` for formatting
- Run \`go vet ./...\` before committing
- Use \`golangci-lint\` for additional linting

#### Required Patterns

**1. Logging with \`slog\`**

Always use \`log/slog\` for structured logging:

\`\`\`go
import "log/slog"

// Good
slog.Info("user created", "user_id", userID, "email", email)
slog.Error("database error", "error", err, "operation", "CreateUser")

// Bad - don't use fmt.Println or log package
fmt.Println("user created:", userID)
log.Printf("error: %v", err)
\`\`\`

**2. Error Handling with \`apperror\`**

NEVER return raw errors to clients. Always use the \`apperror\` package:

\`\`\`go
import "github.com/sergot/tibiacores/backend/pkg/apperror"

// Good - database error
if err != nil {
    return apperror.DatabaseError("failed to create user", err).
        WithDetails(&apperror.DatabaseErrorDetails{
            Operation: "CreateUser",
            Table:     "users",
        })
}

// Good - validation error
if email == "" {
    return apperror.ValidationError("email is required").
        WithDetails(&apperror.ValidationErrorDetails{
            Field:  "email",
            Reason: "cannot be empty",
        })
}

// Good - external service error
if err != nil {
    return apperror.ExternalServiceError("failed to send email", err).
        WithDetails(&apperror.ExternalServiceErrorDetails{
            Service:   "Mailgun",
            Operation: "SendVerificationEmail",
        })
}

// Bad - don't return raw errors
if err != nil {
    return err  // ❌ Never do this
}
\`\`\`

**Error Types Available:**
- \`ValidationError\` → 400 Bad Request
- \`AuthenticationError\` → 401 Unauthorized  
- \`AuthorizationError\` → 403 Forbidden
- \`NotFoundError\` → 404 Not Found
- \`DatabaseError\` → 500 Internal Server Error
- \`InternalServerError\` → 500 Internal Server Error
- \`ExternalServiceError\` → 502 Bad Gateway

**3. Context Passing**

ALWAYS pass context to database operations:

\`\`\`go
// Good
user, err := store.GetUser(ctx, userID)
characters, err := store.ListCharacters(ctx, userID)

// Bad
user, err := store.GetUser(userID)  // ❌ Missing context
\`\`\`

**4. Database Changes Workflow**

When modifying database schema or queries:

1. **Create migration**:
   \`\`\`bash
   cd backend
   goose -dir db/migrations create add_your_feature sql
   \`\`\`

2. **Write migration** in \`db/migrations/YYYYMMDDHHMMSS_add_your_feature.sql\`:
   \`\`\`sql
   -- +goose Up
   -- +goose StatementBegin
   ALTER TABLE users ADD COLUMN new_field TEXT;
   -- +goose StatementEnd

   -- +goose Down
   -- +goose StatementBegin
   ALTER TABLE users DROP COLUMN new_field;
   -- +goose StatementEnd
   \`\`\`

3. **Apply migration**:
   \`\`\`bash
   goose -dir db/migrations postgres "$DATABASE_URL" up
   \`\`\`

4. **Add/update SQL queries** in \`db/queries/*.sql\`:
   \`\`\`sql
   -- name: GetUserWithNewField :one
   SELECT id, email, new_field FROM users WHERE id = $1;
   \`\`\`

5. **Generate Go code**:
   \`\`\`bash
   sqlc generate
   \`\`\`

6. **Use in handler**:
   \`\`\`go
   user, err := h.store.GetUserWithNewField(ctx, userID)
   \`\`\`

For complete documentation see [docs/database.md](docs/database.md) and [docs/setup.md](docs/setup.md).

## Pull Request Process

### Before Submitting

1. **Create tests** for new functionality
2. **Run tests**: \`cd backend && go test ./...\`
3. **Run linters**: \`cd backend && go vet ./...\`
4. **Update documentation** if needed
5. **Regenerate sqlc** if you modified queries: \`sqlc generate\`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

\`\`\`
<type>(<scope>): <description>
\`\`\`

**Examples:**
\`\`\`
feat(auth): add OAuth2 support for Discord
fix(lists): prevent duplicate soul core entries
docs(api): add endpoint documentation for chat
\`\`\`

### Code Review Checklist

Reviewers should verify:

- [ ] Code follows project conventions
- [ ] \`apperror\` used for all error handling
- [ ] \`slog\` used for all logging
- [ ] Context passed to all DB operations
- [ ] Tests cover new functionality
- [ ] No raw SQL queries (only via sqlc)
- [ ] HTTP clients have timeout configuration
- [ ] Documentation updated if needed

## Security Guidelines

1. **Never commit secrets**
2. **Use \`apperror\`** to avoid leaking internal errors
3. **Validate all inputs**
4. **Use sqlc** to prevent SQL injection
5. **Configure timeouts** for external HTTP clients
6. **Hash passwords** with bcrypt

## License

By contributing, you agree that your contributions will be licensed under the project's license (see [LICENSE](LICENSE)).

## Related Documentation

- [docs/setup.md](docs/setup.md) - Detailed development setup
- [docs/database.md](docs/database.md) - Database schema and migrations
