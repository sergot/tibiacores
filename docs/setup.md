# Development Setup Guide

This guide will help you set up TibiaCores for local development.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Docker** and **Docker Compose** (v2.0+) - For running PostgreSQL and the development environment
- **Go** (1.25+) - For backend development
- **Node.js** (20+) and **npm** - For frontend development
- **Make** (optional) - For using Makefile commands

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/sergot/tibiacores.git
cd tibiacores
```

### 2. Environment Configuration

Copy the example environment file and configure your variables:

```bash
cp .env.example .env
```

Edit `.env` and set the required values. At minimum, you need:

- `JWT_SECRET` - Can be any string for development (e.g., `dev-secret-key`)
- `DATABASE_URL` - Default Docker Compose value works: `postgresql://postgres:postgres@localhost:5432/tibiacores?sslmode=disable`
- `FRONTEND_URL` - Default: `http://localhost:5173`

**Optional but recommended for full functionality:**
- Mailgun credentials (for email verification)
- EmailOctopus credentials (for newsletter)
- OAuth credentials (for Discord/Google login)

See [.env.example](../.env.example) for detailed configuration options.

### 3. Start the Development Environment

Using Docker Compose (recommended):

```bash
docker compose up -d
```

This will start:
- PostgreSQL database on port `5432`
- Backend API on port `8080`
- Frontend development server on port `5173`

### 4. Access the Application

Open your browser and navigate to:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080/api/health

## Manual Setup (Without Docker)

### Backend Setup

1. **Start PostgreSQL** (if not using Docker):

```bash
# Using Homebrew (macOS)
brew install postgresql
brew services start postgresql

# Create database
createdb tibiacores
```

2. **Run database migrations**:

```bash
cd backend
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir db/migrations postgres "postgresql://postgres:postgres@localhost:5432/tibiacores?sslmode=disable" up
```

3. **Install Go dependencies**:

```bash
cd backend
go mod download
```

4. **Generate sqlc code** (required after schema changes):

```bash
cd backend
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

5. **Run the backend**:

```bash
cd backend
go run cmd/server/main.go
```

The API will be available at http://localhost:8080

### Frontend Setup

1. **Install dependencies**:

```bash
cd frontend
npm install
```

2. **Run the development server**:

```bash
npm run dev
```

The frontend will be available at http://localhost:5173

## OAuth Provider Setup

To enable OAuth authentication, you need to create applications with the providers:

### Discord OAuth

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Give it a name (e.g., "TibiaCores Local")
4. Navigate to "OAuth2" section
5. Add redirect URI: `http://localhost:5173/auth/discord/callback`
6. Copy "Client ID" and "Client Secret" to your `.env` file:
   ```
   DISCORD_CLIENT_ID=your-client-id
   DISCORD_CLIENT_SECRET=your-client-secret
   DISCORD_REDIRECT_URI=http://localhost:5173/auth/discord/callback
   ```

### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable "Google+ API"
4. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
5. Choose "Web application"
6. Add authorized redirect URI: `http://localhost:5173/auth/google/callback`
7. Copy "Client ID" and "Client Secret" to your `.env` file:
   ```
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URI=http://localhost:5173/auth/google/callback
   ```

## Email Service Setup

### Mailgun Configuration

**Important**: The backend is hardcoded to use **Mailgun EU region** (`mailgun.APIBaseEU`). If you have US region credentials, they will not work.

1. Sign up at [Mailgun](https://www.mailgun.com/)
2. Verify your domain (or use sandbox domain for testing)
3. Get your API key from the dashboard
4. Add to `.env`:
   ```
   MAILGUN_DOMAIN=mg.yourdomain.com
   MAILGUN_API_KEY=your-api-key
   EMAIL_FROM_ADDRESS=noreply@yourdomain.com
   ```

**Troubleshooting**: If you get authentication errors, verify you're using EU region credentials.

### EmailOctopus Configuration

1. Sign up at [EmailOctopus](https://emailoctopus.com/)
2. Create a mailing list
3. Get your API key from account settings
4. Add to `.env`:
   ```
   EMAILOCTOPUS_API_KEY=your-api-key
   EMAILOCTOPUS_LIST_ID=your-list-id
   ```

## Database Management

### Running Migrations

Migrations are located in `backend/db/migrations/`. To run them:

```bash
cd backend
goose -dir db/migrations postgres "$DATABASE_URL" up
```

To rollback the last migration:

```bash
goose -dir db/migrations postgres "$DATABASE_URL" down
```

To check migration status:

```bash
goose -dir db/migrations postgres "$DATABASE_URL" status
```

### Seeding Initial Data

The creatures data is loaded from `data/creatures.txt`. To populate the database:

```bash
# This should be done automatically on first run
# Or you can manually insert using psql:
psql $DATABASE_URL -c "COPY creatures(name) FROM '/path/to/data/creatures.txt';"
```

### Regenerating sqlc Code

After modifying SQL queries in `backend/db/queries/*.sql`, regenerate the Go code:

```bash
cd backend
sqlc generate
```

This will update files in `backend/db/sqlc/`.

## Common Development Tasks

### Running Tests

Backend tests:
```bash
cd backend
go test ./...
```

Frontend tests:
```bash
cd frontend
npm run test
```

### Code Linting

Backend:
```bash
cd backend
go vet ./...
golangci-lint run
```

Frontend:
```bash
cd frontend
npm run lint
```

### Building for Production

Backend:
```bash
cd backend
go build -o bin/server cmd/server/main.go
```

Frontend:
```bash
cd frontend
npm run build
```

## Troubleshooting

### Database Connection Errors

**Problem**: `unable to connect to database`

**Solution**:
- Verify PostgreSQL is running: `docker compose ps` or `brew services list`
- Check `DATABASE_URL` in `.env` matches your database configuration
- Ensure port 5432 is not blocked by firewall

### Backend Fails to Start

**Problem**: `JWT_SECRET not set`

**Solution**: Ensure `.env` file exists in project root with `JWT_SECRET` defined

**Problem**: `Error initializing email service`

**Solution**: Mailgun credentials are invalid or missing. Either:
- Add valid Mailgun credentials to `.env`, or
- Comment out email service initialization if not needed for development

### Frontend Cannot Connect to Backend

**Problem**: CORS errors in browser console

**Solution**:
- Verify backend is running on port 8080
- Ensure `FRONTEND_URL=http://localhost:5173` is set in backend `.env`
- Check CORS configuration in `backend/cmd/server/main.go`

### OAuth Callback Fails

**Problem**: `invalid state parameter` or redirect errors

**Solution**:
- Verify redirect URIs match exactly in provider settings and `.env`
- Clear browser cookies and try again
- Check that provider credentials are correct

### Email Verification Not Working

**Problem**: Verification emails not sent

**Solution**:
- Check Mailgun credentials and region (must be EU)
- Verify `EMAIL_FROM_ADDRESS` is authorized in Mailgun
- Check backend logs for email service errors: `docker compose logs backend`

## Development Workflow

1. **Create a feature branch**: `git checkout -b feature/your-feature`
2. **Make changes** to code
3. **Run migrations** if you modified database schema
4. **Regenerate sqlc** if you modified SQL queries
5. **Run tests**: `go test ./...` and `npm run test`
6. **Lint code**: `go vet ./...` and `npm run lint`
7. **Commit changes**: Follow [conventional commits](https://www.conventionalcommits.org/)
8. **Push and create PR**

## Next Steps

- Review the [API Reference](api-reference.md) to understand available endpoints
- Read the [Database Schema](database.md) to understand the data model
- Check [CONTRIBUTING.md](../CONTRIBUTING.md) for coding standards
- Explore [Architecture Documentation](architecture.md) for system design

## Related Documentation

- [.env.example](../.env.example) - Complete environment variable reference
- [database.md](database.md) - Database schema and migrations
- [deployment.md](deployment.md) - Production deployment guide
- [troubleshooting.md](troubleshooting.md) - Common issues and solutions
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
