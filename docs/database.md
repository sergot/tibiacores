# Database Schema Documentation

This document provides comprehensive information about the TibiaCores database schema, migrations, and data management workflows.

## Overview

TibiaCores uses **PostgreSQL 17** as its database with the following tools and patterns:

- **Migration Tool**: [Goose](https://github.com/pressly/goose) for schema versioning
- **Query Layer**: [sqlc](https://sqlc.dev/) for type-safe Go code generation from SQL queries
- **Connection Pool**: `pgx/v5` with `pgxpool` for connection management
- **Schema Location**: `backend/db/migrations/`
- **Queries Location**: `backend/db/queries/`

## Entity Relationship Diagram

```mermaid
erDiagram
    users ||--o{ characters : owns
    users ||--o{ lists : creates
    users ||--o{ lists_users : joins
    users ||--o{ character_claims : claims
    users ||--o{ list_chat_messages : sends
    users ||--o{ list_user_read_status : tracks
    
    characters ||--o{ character_claims : "claimed via"
    characters ||--o{ lists_users : "participates with"
    characters ||--o{ characters_soulcores : unlocks
    characters ||--o{ character_soulcore_suggestions : receives
    characters ||--o{ list_chat_messages : "sent by"
    
    lists ||--o{ lists_users : "has members"
    lists ||--o{ lists_soulcores : contains
    lists ||--o{ character_soulcore_suggestions : generates
    lists ||--o{ list_chat_messages : "has chat"
    lists ||--o{ list_user_read_status : "has read status"
    
    creatures ||--o{ lists_soulcores : "tracked in"
    creatures ||--o{ characters_soulcores : "unlocked by"
    creatures ||--o{ character_soulcore_suggestions : suggested
    
    users {
        uuid id PK
        boolean is_anonymous
        text email UK
        text password
        boolean email_verified
        uuid email_verification_token
        timestamptz email_verification_expires_at
        timestamptz created_at
        timestamptz updated_at
    }
    
    characters {
        uuid id PK
        uuid user_id FK
        text name UK
        text world
        timestamptz created_at
        timestamptz updated_at
    }
    
    character_claims {
        uuid id PK
        uuid character_id FK
        uuid claimer_id FK
        text verification_code
        text status "pending|approved|rejected"
        timestamptz last_checked_at
        timestamptz created_at
        timestamptz updated_at
    }
    
    lists {
        uuid id PK
        uuid author_id FK
        text name
        uuid share_code UK
        text world
        timestamptz created_at
        timestamptz updated_at
    }
    
    lists_users {
        uuid list_id PK_FK
        uuid user_id PK_FK
        uuid character_id PK_FK
        boolean active
    }
    
    creatures {
        uuid id PK
        text name UK
        integer difficulty "1-5"
    }
    
    lists_soulcores {
        uuid list_id PK_FK
        uuid creature_id PK_FK
        uuid added_by_user_id FK
        soulcore_status status "obtained|unlocked"
    }
    
    characters_soulcores {
        uuid character_id PK_FK
        uuid creature_id PK_FK
        timestamptz created_at
    }
    
    character_soulcore_suggestions {
        uuid character_id PK_FK
        uuid creature_id PK_FK
        uuid list_id FK
        timestamptz suggested_at
    }
    
    list_chat_messages {
        uuid id PK
        uuid list_id FK
        uuid user_id FK
        uuid character_id FK
        text message
        timestamptz created_at
    }
    
    list_user_read_status {
        uuid user_id PK_FK
        uuid list_id PK_FK
        timestamptz last_read_at
    }
```

## Tables Reference

### Core Tables

#### users
Stores user accounts, supporting both anonymous and registered users.

**Columns:**
- `id` (UUID, PK) - Unique user identifier
- `is_anonymous` (BOOLEAN) - Whether user is anonymous or registered
- `email` (TEXT, UNIQUE) - User email (NULL for anonymous users)
- `password` (TEXT) - Bcrypt hashed password (NULL for OAuth/anonymous)
- `email_verified` (BOOLEAN) - Email verification status
- `email_verification_token` (UUID) - Token for email verification
- `email_verification_expires_at` (TIMESTAMPTZ) - Verification token expiry
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

**Constraints:**
- `email_unique_if_not_anonymous`: Ensures anonymous users have NULL email, registered users have email

**Indexes:**
- `idx_users_email` on `email` (partial index where email IS NOT NULL)

**Design Notes:**
- Anonymous users are auto-created when accessing certain endpoints without auth
- Anonymous users can be "upgraded" to registered accounts via signup or OAuth
- Email is unique across all registered users but NULL for all anonymous users

---

#### characters
Represents Tibia game characters linked to users.

**Columns:**
- `id` (UUID, PK) - Unique character identifier
- `user_id` (UUID, FK → users) - Owner of the character
- `name` (TEXT, UNIQUE) - Character name (must match Tibia.com)
- `world` (TEXT) - Tibia game world (e.g., "Antica", "Secura")
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

**Constraints:**
- Character name is globally unique (one character = one owner at a time)
- Character can be transferred via the claims system

**Design Notes:**
- Characters are verified via TibiaData API to ensure they exist
- Character ownership can be disputed and transferred using `character_claims`

---

#### character_claims
Handles character ownership verification and transfers.

**Columns:**
- `id` (UUID, PK)
- `character_id` (UUID, FK → characters) - Character being claimed
- `claimer_id` (UUID, FK → users) - User attempting to claim
- `verification_code` (TEXT) - Code to be added to character comment on Tibia.com
- `status` (TEXT) - `pending` | `approved` | `rejected`
- `last_checked_at` (TIMESTAMPTZ) - Last time TibiaData API was checked
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

**Workflow:**
1. User initiates claim → backend generates `verification_code`
2. User adds code to character comment on Tibia.com
3. Background job polls TibiaData API every 15 minutes
4. If code matches → status becomes `approved`, character ownership transfers
5. Claims expire after 24 hours if not verified

**Design Notes:**
- When a claim is approved, previous owner's list memberships are deactivated (`lists_users.active = false`)
- Multiple pending claims for same character can exist (first verified wins)

---

#### lists
Collaborative soul core tracking lists tied to specific Tibia worlds.

**Columns:**
- `id` (UUID, PK)
- `author_id` (UUID, FK → users) - List creator
- `name` (TEXT) - List name
- `share_code` (UUID, UNIQUE) - Used for invite links
- `world` (TEXT) - Tibia world (must match members' characters)
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

**Design Notes:**
- `share_code` is publicly shareable for joining lists
- All list members must have characters from the same world
- Author has special permissions (can remove members, delete list)

---

#### lists_users
Junction table for list memberships.

**Columns:**
- `list_id` (UUID, PK/FK → lists)
- `user_id` (UUID, PK/FK → users)
- `character_id` (UUID, PK/FK → characters)
- `active` (BOOLEAN) - Membership status

**Composite Primary Key:** `(list_id, user_id, character_id)`

**Design Notes:**
- `active` flag is set to `false` when character ownership changes (via claims)
- Allows tracking historical memberships without data loss
- A user can join same list with multiple characters
- Queries must filter by `active = true` for current members

---

#### creatures
Catalog of all soul core creatures in Tibia.

**Columns:**
- `id` (UUID, PK)
- `name` (TEXT, UNIQUE) - Creature name
- `difficulty` (INTEGER) - Difficulty rating (1-5)
  - 1: Easy
  - 2: Medium
  - 3: Hard
  - 4: Very Hard
  - 5: Extreme (rare)

**Design Notes:**
- Pre-populated from `data/creatures.txt` (800+ creatures)
- Difficulty added in migration `20250324000002`
- Serves as reference data, rarely changes

---

### Soul Core Tracking Tables

#### lists_soulcores
Soul cores tracked within a list.

**Columns:**
- `list_id` (UUID, PK/FK → lists)
- `creature_id` (UUID, PK/FK → creatures)
- `added_by_user_id` (UUID, FK → users) - Who added this core
- `status` (soulcore_status) - `obtained` | `unlocked`

**Composite Primary Key:** `(list_id, creature_id)`

**Custom Type:** `soulcore_status` enum

**Status Workflow:**
- **obtained**: Core is tracked but not yet unlocked
- **unlocked**: Core has been fully unlocked by a list member
  - When status changes to `unlocked`, suggestions are created for other members

**Design Notes:**
- Only adder or list author can modify/remove core
- Tracks who added each core for permission checks

---

#### characters_soulcores
Individual character's unlocked soul cores.

**Columns:**
- `character_id` (UUID, PK/FK → characters)
- `creature_id` (UUID, PK/FK → creatures)
- `created_at` (TIMESTAMPTZ)

**Composite Primary Key:** `(character_id, creature_id)`

**Design Notes:**
- Automatically populated when user marks core as unlocked in a list
- Also used for public character profiles
- No status field (all entries are "unlocked")

---

#### character_soulcore_suggestions
Smart suggestions for characters to hunt specific cores.

**Columns:**
- `character_id` (UUID, PK/FK → characters)
- `creature_id` (UUID, PK/FK → creatures)
- `list_id` (UUID, FK → lists) - Source list that generated suggestion
- `suggested_at` (TIMESTAMPTZ)

**Composite Primary Key:** `(character_id, creature_id)`

**Indexes:**
- `character_soulcore_suggestions_character_id_idx` on `character_id`

**Generation Logic:**
When a list member marks a core as "unlocked":
1. Check all other list members
2. For each member, if they don't have this core yet → create suggestion
3. User can accept (adds to their character) or dismiss (deletes suggestion)

**Design Notes:**
- Prevents duplicate suggestions (composite PK)
- User sees "pending suggestions" count in UI
- Reduces manual tracking overhead for teams

---

### Chat Tables

#### list_chat_messages
Messages in list-specific chat channels.

**Columns:**
- `id` (UUID, PK)
- `list_id` (UUID, FK → lists)
- `user_id` (UUID, FK → users)
- `character_id` (UUID, FK → characters) - Which character sent the message
- `message` (TEXT)
- `created_at` (TIMESTAMPTZ)

**Indexes:**
- `idx_list_chat_messages_list_id` on `list_id`
- `idx_list_chat_messages_created_at` on `created_at`

**Design Notes:**
- Messages are tied to characters (shows which character is speaking)
- Only message author can delete their own messages
- Supports pagination via `created_at` ordering

---

#### list_user_read_status
Tracks read receipts for chat messages.

**Columns:**
- `user_id` (UUID, PK/FK → users)
- `list_id` (UUID, PK/FK → lists)
- `last_read_at` (TIMESTAMPTZ)

**Composite Primary Key:** `(user_id, list_id)`

**Design Notes:**
- Updated when user views chat or marks messages as read
- Used to calculate unread message counts
- Counts messages where `created_at > last_read_at`

---

## Database Migrations

### Migration Tool: Goose

Migrations are located in `backend/db/migrations/` and use Goose syntax.

**Migration File Format:**
```
YYYYMMDDHHMMSS_description.sql
```

**Structure:**
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE ...
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ...
-- +goose StatementEnd
```

### Running Migrations

**Apply all pending migrations:**
```bash
cd backend
goose -dir db/migrations postgres "$DATABASE_URL" up
```

**Rollback last migration:**
```bash
goose -dir db/migrations postgres "$DATABASE_URL" down
```

**Check migration status:**
```bash
goose -dir db/migrations postgres "$DATABASE_URL" status
```

**Create new migration:**
```bash
goose -dir db/migrations create add_feature_name sql
```

### Migration History

| Migration | Description |
|-----------|-------------|
| `20250317101323_init_creatures.sql` | Create creatures table, seed with 800+ creatures |
| `20250317103543_init_users.sql` | Create users table with anonymous user support |
| `20250317103641_init_lists.sql` | Create lists table |
| `20250317103747_init_characters.sql` | Create characters and character_claims tables |
| `20250317103842_init_lists_users.sql` | Create list membership junction table |
| `20250317103920_init_lists_soulcores.sql` | Create list soul cores tracking with status enum |
| `20250317103951_init_characters_soulcores.sql` | Create character soul cores table |
| `20250324000001_add_soulcore_suggestions.sql` | Add smart suggestions system |
| `20250324000002_add_creature_difficulty.sql` | Add difficulty ratings to creatures |
| `20250520000001_add_chat_messages.sql` | Add list chat messaging |
| `20250520000002_add_chat_read_status.sql` | Add chat read receipts |

---

## sqlc Code Generation

### What is sqlc?

`sqlc` generates type-safe Go code from SQL queries. This provides:
- **Compile-time query validation** - Syntax errors caught before runtime
- **Type safety** - Query results mapped to Go structs
- **SQL injection prevention** - Parameterized queries only
- **No ORM magic** - Write SQL, get Go functions

### Configuration

Located in `backend/sqlc.yaml`:
```yaml
version: "2"
sql:
  - schema: "db/migrations"
    queries: "db/queries"
    engine: "postgresql"
    gen:
      go:
        package: "sqlc"
        out: "db/sqlc"
        emit_json_tags: true
        emit_empty_slices: true
```

### Workflow

1. **Write SQL queries** in `backend/db/queries/*.sql`:
```sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;
```

2. **Generate Go code**:
```bash
cd backend
sqlc generate
```

3. **Use generated functions** in handlers:
```go
user, err := store.GetUserByEmail(ctx, email)
```

### Query File Organization

- `users.sql` - User and authentication queries
- `characters.sql` - Character and claims queries
- `lists.sql` - List management queries
- `chat.sql` - Chat message queries
- `creatures.sql` - Creature catalog queries
- `suggestions.sql` - Suggestion system queries

### Adding New Queries

1. Add SQL to appropriate file in `db/queries/`
2. Run `sqlc generate`
3. Use new function from `db/sqlc/` package
4. If schema changed, create migration first

---

## Schema Modification Workflow

When adding a new feature that requires database changes:

### 1. Create Migration

```bash
cd backend
goose -dir db/migrations create your_feature_name sql
```

Edit the generated file:
```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN new_field TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN new_field;
-- +goose StatementEnd
```

### 2. Apply Migration

```bash
goose -dir db/migrations postgres "$DATABASE_URL" up
```

### 3. Add/Update Queries

Edit `db/queries/*.sql`:
```sql
-- name: GetUserWithNewField :one
SELECT id, email, new_field FROM users WHERE id = $1;
```

### 4. Generate sqlc Code

```bash
sqlc generate
```

### 5. Update Handlers

Use new queries in `handlers/`:
```go
user, err := h.store.GetUserWithNewField(ctx, userID)
```

### 6. Test

Write tests for new functionality:
```go
func TestNewFeature(t *testing.T) {
    // Use mock store generated by gomock
    mockStore.EXPECT().GetUserWithNewField(...)
}
```

---

## Key Design Decisions

### Why Composite Primary Keys?

**`lists_users (list_id, user_id, character_id)`**
- Prevents duplicate memberships
- Natural unique constraint
- Efficient for junction table queries

**`character_soulcore_suggestions (character_id, creature_id)`**
- Prevents duplicate suggestions
- No need for separate unique index

**Trade-off**: Harder to reference in other foreign keys (but not needed here)

---

### Why Active Flag on lists_users?

When a character ownership changes via claims:
- Setting `active = false` preserves historical data
- Shows who was in the list previously
- Allows "reactivation" if character is re-claimed by original owner
- Alternative would be deletion (loses history)

**Queries must filter**: `WHERE active = true`

---

### Why Character Name Unique Constraint?

- One character can only belong to one user at a time
- Matches real Tibia game mechanics (character names are globally unique)
- Ownership transfer handled explicitly via claims system
- Prevents race conditions and duplicate characters

**Edge case**: If character is deleted in Tibia, name remains locked in database until manual cleanup

---

### Why Anonymous User Model?

- Reduces friction (users can try features without signup)
- Seamless upgrade to registered account (preserves characters/lists)
- JWT still issued (`has_email: false`)
- Challenges: orphaned anonymous accounts, no way to contact users

**Constraint**: Anonymous users have `email = NULL`, registered users must have email

---

### Why soulcore_status Enum?

- Type safety at database level
- Only two valid states: `obtained` | `unlocked`
- PostgreSQL enums are efficient (stored as integers)
- Prevents typos and invalid states

**Adding new statuses** requires migration to alter enum type

---

## Performance Considerations

### Indexes

Current indexes optimize common query patterns:
- `idx_users_email` - Email lookups for authentication
- `idx_list_chat_messages_list_id` - Chat history retrieval
- `idx_list_chat_messages_created_at` - Chronological ordering
- `character_soulcore_suggestions_character_id_idx` - Pending suggestions lookup

### Connection Pooling

`pgxpool` configuration in `cmd/server/main.go`:
- Handles connection lifecycle
- Prevents connection exhaustion
- Automatic reconnection on failures

### Query Patterns

- All queries use parameterized placeholders (`$1`, `$2`) via sqlc
- Context passed to all database operations for cancellation
- Transactions used for multi-step operations (claim approval, suggestion generation)

---

## Backup and Maintenance

### Database Backups

**Recommended strategy:**
```bash
# Full backup
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql

# Restore
psql $DATABASE_URL < backup_20260203.sql
```

### Cleanup Tasks

**Expired verification tokens:**
```sql
DELETE FROM users 
WHERE email_verification_expires_at < NOW() 
AND email_verified = false;
```

**Old pending claims (>24 hours):**
```sql
DELETE FROM character_claims 
WHERE status = 'pending' 
AND created_at < NOW() - INTERVAL '24 hours';
```

**Orphaned anonymous users (no characters, no lists, >30 days old):**
```sql
DELETE FROM users
WHERE is_anonymous = true
AND created_at < NOW() - INTERVAL '30 days'
AND id NOT IN (SELECT user_id FROM characters)
AND id NOT IN (SELECT author_id FROM lists)
AND id NOT IN (SELECT user_id FROM lists_users);
```

---

## Related Documentation

- [setup.md](setup.md) - Database setup and migration workflow
- [api-reference.md](api-reference.md) - API endpoints that interact with database
- [CONTRIBUTING.md](../CONTRIBUTING.md) - How to add new features requiring schema changes
- [architecture.md](architecture.md) - Overall system architecture including database layer
