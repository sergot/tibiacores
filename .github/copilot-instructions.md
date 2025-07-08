# TibiaCores Copilot Instructions

This document provides an overview of the TibiaCores project structure and key concepts to help GitHub Copilot better understand and assist with the codebase.

## Project Overview

TibiaCores is a web application for tracking and managing soul cores in the game Tibia. It allows users to create lists, collaborate with others, and track progress of their soul core collections.

## Environment Setup

The project uses Docker Compose to run both the backend and frontend services. This simplifies development and ensures consistent environments across different development machines.

- **Docker Compose**: Used to orchestrate and run both frontend and backend services
- **Development**: `docker compose up` starts the development environment
- **Production**: Production deployment also utilizes Docker Compose with production-specific configurations

## Tech Stack

### Backend
- **Language**: Go
- **Framework**: Echo (web framework)
- **Database**: PostgreSQL with pgx driver
- **Schema Migration**: Goose
- **Database Query**: SQLC for type-safe SQL

### Frontend
- **Language**: TypeScript
- **Framework**: Vue 3 (with Composition API)
- **UI**: TailwindCSS
- **Internationalization**: Vue I18n
- **HTTP Client**: Axios

## Key Concepts

### Lists
Lists are collections of soul cores that users can create and share with others. Each list belongs to an author and can have multiple members.

### Characters
Users can claim and verify their in-game characters, which are then used to interact with lists.

### Soul Cores
Soul cores represent in-game items that players can collect. They have two states:
- **Obtained**: The player has the soul core item
- **Unlocked**: The player has unlocked the creature's ability in the soulcatcher

### Chat
Each list has a chat feature where members can communicate with each other. Chat messages are linked to a specific character belonging to the user. The chat implementation includes:
- Floating chat bubble interface in the bottom right corner
- Unread message notifications
- Real-time message polling when chat is open
- Auto-scrolling to the latest messages
- Message sent/received indicators
- Character-based messaging (users send messages as their characters)

## Database Schema

### Key Tables
- `users`: User accounts
- `characters`: In-game characters claimed by users
- `lists`: Soul core lists created by users
- `lists_users`: Many-to-many relationship between lists and users (via characters)
- `lists_soulcores`: Soul cores in a specific list
- `characters_soulcores`: Soul cores owned by characters
- `list_chat_messages`: Chat messages in a list

## Error Handling System

The application uses a sophisticated custom error handling system located in `backend/pkg/apperror/`.

### AppError Structure
All application errors use the `AppError` type which provides:
- **Type-safe error categorization**: `validation`, `authorization`, `not_found`, `database`, `internal`, `external`
- **Structured error details**: Type-safe details for different error types
- **Request context tracking**: Function, file, line, timestamp, operation, request ID, user ID
- **Client-safe responses**: Automatic filtering of sensitive information for API responses
- **Comprehensive logging**: Structured JSON logging with full error context

### Error Types and Details
- **ValidationErrorDetails**: Field-level validation errors with field, value, and reason
- **DatabaseErrorDetails**: Database operation errors with operation, table, and optional query
- **ExternalServiceErrorDetails**: Third-party service errors with service, operation, and endpoint
- **AuthorizationErrorDetails**: Authorization failures with reason and optional field

### Usage Patterns
```go
// Create specific error types
return apperror.ValidationError("Invalid email format", nil).
    WithDetails(&apperror.ValidationErrorDetails{
        Field: "email",
        Value: email,
        Reason: "must be a valid email address",
    })

// Database errors with context
return apperror.DatabaseError("Failed to create user", err).
    WithDetails(&apperror.DatabaseErrorDetails{
        Operation: "INSERT",
        Table: "users",
    })
```

### Error Middleware
The `ErrorHandler` in `middleware/error_middleware.go`:
- Automatically adds request context (request ID, user ID, operation)
- Converts all error types to consistent JSON responses
- Provides comprehensive logging
- Includes panic recovery with stack traces
- Handles JSON serialization failures gracefully

## Testing Patterns

The backend follows consistent testing patterns using gomock for mocking and testify for assertions.

### Test Structure
All handler tests follow this pattern:
```go
func TestHandlerName(t *testing.T) {
    testCases := []struct {
        name          string
        setupRequest  func(c echo.Context)
        setupMocks    func(store *mockdb.MockStore, /* other dependencies */)
        expectedCode  int
        expectedError string
        checkResponse func(t *testing.T, response ResponseType)
    }{
        // Test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Setup gomock controller
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            // Create mocks and handler
            store := mockdb.NewMockStore(ctrl)
            // Setup HTTP request/response
            // Execute test
            // Verify results
        })
    }
}
```

### Mock Database Patterns
- **Generated mocks**: All database mocks are generated using `gomock` from the SQLC `Store` interface
- **Expectation patterns**: Use `store.EXPECT().MethodName(gomock.Any(), params).Return(result, error)`
- **Parameter matching**: Use `gomock.Any()` for context, specific values or matchers for other parameters
- **Order independence**: Tests don't rely on call order unless specifically needed

### Common Test Scenarios
Every handler typically tests:
- **Success case**: Happy path with valid input
- **Authentication errors**: Invalid user ID, missing authentication
- **Authorization errors**: Access to resources owned by other users
- **Validation errors**: Invalid input parameters
- **Database errors**: Simulated database failures
- **Not found cases**: Resources that don't exist

### Helper Functions
- `MustHashPassword(password string)`: Test helper for password hashing
- `newMockEmailService(ctrl *gomock.Controller)`: Creates email service mocks
- Mock services implement interfaces to enable easy testing

## Database Query Patterns with SQLC

The application uses SQLC for type-safe database queries.

### SQLC Configuration
- **Config file**: `backend/sqlc.yaml`
- **Query files**: `backend/db/queries/*.sql`
- **Generated code**: `backend/db/sqlc/`
- **Models**: All database types generated in `models.go`

### Query Organization
Queries are organized by domain:
- `users.sql`: User management queries
- `characters.sql`: Character-related operations
- `lists.sql`: List management
- `chat.sql`: Chat message operations
- `creatures.sql`: Creature data queries
- `suggestions.sql`: Soulcore suggestion queries

### SQLC Patterns
- **Parameter structs**: Complex queries use parameter structs (e.g., `CreateUserParams`)
- **Return types**: Queries return either single structs or row slices
- **Custom return types**: Complex queries generate custom row types (e.g., `GetListMembersRow`)
- **UUID handling**: Consistent use of `github.com/google/uuid` for UUID types
- **Nullable types**: Uses `pgtype` for nullable database fields

### Store Interface
The `Store` interface in `db/sqlc/store.go`:
- Embeds the generated `Querier` interface
- Provides a consistent interface for all database operations
- Enables easy mocking for tests
- Supports connection pooling through `pgxpool.Pool`

## Validation Patterns

### Request Validation
- **Struct tags**: Uses Go struct tags for basic validation
- **Custom validation**: Manual validation in handlers for complex business rules
- **Error integration**: Validation errors use the `AppError` system with `ValidationErrorDetails`

### Common Validation Patterns
- **UUID validation**: Check for valid UUID format in path parameters
- **User ownership**: Verify users can only access their own resources
- **Required fields**: Check for required request body fields
- **Business logic**: Validate business rules (e.g., character belongs to correct world)

## Middleware Architecture

### Custom Middleware
The application includes several custom middleware components:

#### Error Middleware
- **Global error handling**: Catches all errors and converts to consistent JSON responses
- **Request context**: Adds request ID, user ID, and operation to error context
- **Panic recovery**: Graceful handling of panics with stack trace logging
- **Client safety**: Filters sensitive information from error responses

#### Authentication Middleware
- **JWT validation**: Validates JWT tokens and extracts user information
- **Context injection**: Adds user ID to request context for handlers
- **Protected routes**: Applied to routes requiring authentication

### Middleware Patterns
- **Context usage**: Middleware stores data in Echo context using `c.Set()`
- **Error propagation**: Middleware returns errors that are handled by error middleware
- **Request/response modification**: Middleware can modify requests and responses
- **Chaining**: Multiple middleware can be chained together

## Directory Structure

### Backend
- `auth/`: Authentication logic including JWT and OAuth
- `cmd/server/`: Application entry point
- `db/migrations/`: Database schema migrations
- `db/queries/`: SQL queries (used by SQLC)
- `db/sqlc/`: Generated Go code for database operations
- `handlers/`: HTTP request handlers
- `middleware/`: HTTP middleware components
- `pkg/`: Shared packages and utilities
- `services/`: External services integration

### Frontend
- `src/components/`: Reusable Vue components (ChatWindow, CreatureSelect, etc.)
- `src/views/`: Vue components representing application pages (ListDetailView, CharacterView, etc.)
- `src/stores/`: Pinia stores for state management (user, chatNotifications, etc.)
- `src/services/`: API services and external integrations
- `src/router/`: Vue Router configuration
- `src/i18n/`: Internationalization setup and translations
- `src/assets/`: Static assets like images, styles, and fonts

## Common Development Tasks

### Adding a New Feature
1. Consider if database schema changes are needed (create migration files)
2. Add SQL queries in the appropriate query file
3. Implement backend handlers
4. Create or update frontend components
5. Add translations for any new UI text
6. **Run quality checks**: Always run `golangci-lint run`, `npm run lint`, and `npm run build` before committing

### Database Migrations
When modifying the database schema:
1. Create a new migration file in `backend/db/migrations/` with the prefix format `YYYYMMDD000001_`
2. Run migrations using goose

### Adding New API Endpoints
1. Implement handler function in the appropriate file in `backend/handlers/`
2. Register the route in `backend/cmd/server/main.go`
3. Create frontend service to call the new endpoint

### Creating New Handlers
Follow these patterns:
1. **Error handling**: Use the `AppError` system for all errors
2. **Authentication**: Extract user ID from context: `c.Get("user_id").(string)`
3. **Validation**: Validate input and return appropriate error types
4. **Database operations**: Use the store interface methods
5. **Response format**: Return consistent JSON responses

### Adding Database Queries
1. Write SQL query in appropriate file in `backend/db/queries/`
2. Run `sqlc generate` to generate Go code
3. Use generated methods in handlers
4. Add mocks expectations in tests

### Translation
All user-facing text should be added to the translation files in `frontend/src/i18n/locales/`. 

When adding new text strings:
1. First add them to the English locale file (`en.json`)
2. Then add the translations to all other locale files: 
   - German (`de.json`)
   - Spanish (`es.json`)
   - Polish (`pl.json`) 
   - Portuguese (`pt.json`)
3. Make sure the structure matches exactly between all files
4. Check for completeness across all locales when adding new features
5. Run `npm run check-translations` from the frontend directory to verify there are no missing translations or errors
   - This tool will:
     - Find all missing translations across language files
     - Identify unused translations
     - Clean up translation files by removing unused keys
     - Check for incorrect interpolation formats (`{{variable}}` instead of the correct `{variable}`) as Vue I18n requires this format

## Authentication and Authorization

The application uses JWT for authentication. Protected endpoints require a valid JWT token.
OAuth integration is available for third-party login providers.

## State Management

The application uses Pinia for state management with the following key stores:
- **userStore**: Manages user authentication state and user data
- **chatNotificationsStore**: Tracks unread chat messages and notifications
- **listStore**: Manages active lists and their data
- **charactersStore**: Manages character information

## UI Components and Interactions

- **Responsive Design**: All components are designed to work well on both desktop and mobile devices
- **Animations**: The application uses CSS animations for transitions and to enhance user experience
- **Real-time Updates**: Chat messages and notifications are updated in real-time
- **Modals and Dialogs**: Used for confirmations, sharing features, and detailed information
- **Form Validation**: Client-side validation for all user inputs

## Testing

The backend includes unit tests for handlers and services. Tests use mocks for database operations.

## Code Quality and Testing

Before committing any changes to the codebase, contributors must run the following commands to ensure code quality and prevent build failures:

### Backend Quality Checks
From the `backend/` directory:
```bash
# Run Go linter
golangci-lint run

# Run all Go tests
go test ./...
```

### Frontend Quality Checks
From the `frontend/` directory:
```bash
# Run TypeScript/Vue linter
npm run lint

# Build the frontend to check for compilation errors
npm run build

# Check translations for completeness and consistency
npm run check-translations
```

### Required Tools
- **golangci-lint**: Install following [official instructions](https://golangci-lint.run/usage/install/)
- **Node.js**: Required for frontend development and build processes
- **Docker Compose**: For running the full development environment

### Continuous Integration
These same checks are run in CI/CD pipelines. Failing any of these checks will prevent code from being merged.
