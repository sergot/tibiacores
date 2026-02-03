# Authentication & Security Design

## Overview
TibiaCores uses a modern, stateless authentication system based on JSON Web Tokens (JWT) and OAuth2. The system is designed to be secure, scalable, and easy to integrate with frontend applications.

## Authentication Methods

### 1. Email & Password
- **Registration**: Users sign up with email and password.
- **Hashing**: Passwords are hashed using `bcrypt` with a work factor of 10.
- **Verification**: Email verification flows ensure user authenticity before full account activation.

### 2. OAuth2 Providers
- **Discord**: Link and login with Discord account.
- **Google**: Link and login with Google account.
- **Flow**:
  1. Frontend requests redirect URL from backend.
  2. Backend generates a secure random state and sets a `HttpOnly` cookie.
  3. Frontend redirects user to provider.
  4. User approves and is redirected back to Frontend.
  5. Frontend sends code and state to Backend.
  6. Backend validates state against the cookie (CSRF protection) and exchanges code for user profile.

## Security Mechanisms

### Token Management (JWT)
- **Algorithm**: HMAC-SHA256 (`HS256`).
- **Structure**:
  ```json
  {
    "user_id": "uuid-string",
    "has_email": true,
    "exp": 1234567890,
    "iss": "tibiacores"
  }
  ```
- **Storage**: Tokens are returned to the client and should be stored securely (e.g., in memory or secure storage mechanisms).
- **Transport**: All requests must include the token in the `Authorization` header: `Bearer <token>`.

### CSRF Protection
- **OAuth**: Uses the "Double Submit Cookie" pattern. A random state is set in a secure, HTTP-only cookie and also sent to the OAuth provider. Upon callback, the backend verifies that the state returned by the provider matches the cookie.
- **API**: Since the API relies on Authorization headers (not cookies) for authentication, it is naturally resistant to CSRF attacks for general API endpoints.

### Input Validation
- All inputs are validated using strictly typed Go structs and the `validator` package.
- SQL Injection is prevented by using `sqlc` which generates type-safe code and uses parameterized queries.

## Future Roadmap
- **Refresh Tokens**: Implement refresh tokens to allow shorter access token lifespans.
- **Rate Limiting**: Add rate limiting to sensitive endpoints (`/login`, `/signup`).
- **MFA**: Support Multi-Factor Authentication for enhanced security.
