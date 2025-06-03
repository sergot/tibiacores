# TibiaCores Architecture Diagram

```mermaid
graph TB
    %% Frontend Layer
    subgraph "Frontend (Vue 3 + TypeScript)"
        UI[User Interface]
        subgraph "Vue Components"
            Views[Views<br/>- ListDetailView<br/>- CharacterView<br/>- etc.]
            Components[Components<br/>- ChatWindow<br/>- CreatureSelect<br/>- etc.]
        end
        
        subgraph "State Management"
            Stores[Pinia Stores<br/>- userStore<br/>- chatNotificationsStore<br/>- listStore<br/>- charactersStore]
        end
        
        subgraph "Services & Utils"
            APIServices[API Services<br/>- Axios HTTP Client]
            Router[Vue Router]
            I18n[Vue I18n<br/>Translations]
        end
    end

    %% Backend Layer
    subgraph "Backend (Go + Echo)"
        subgraph "HTTP Layer"
            Routes[Routes & Middleware<br/>- Authentication<br/>- Error Handling<br/>- CORS]
            Handlers[Handlers<br/>- Users<br/>- Characters<br/>- Lists<br/>- Chat<br/>- Creatures]
        end
        
        subgraph "Business Logic"
            Auth[Authentication<br/>- JWT<br/>- OAuth2<br/>- Password Hashing]
            Services[Services<br/>- Email Service<br/>- TibiaData API]
            Validation[Validation<br/>- Request Validation<br/>- Business Rules]
        end
        
        subgraph "Error Handling"
            AppError[AppError System<br/>- Type-safe Errors<br/>- Structured Details<br/>- Context Tracking]
        end
    end

    %% Data Layer
    subgraph "Data Layer"
        subgraph "Database Operations"
            SQLC[SQLC Generated Code<br/>- Type-safe Queries<br/>- Store Interface]
            Queries[SQL Queries<br/>- users.sql<br/>- characters.sql<br/>- lists.sql<br/>- chat.sql<br/>- creatures.sql]
        end
        
        subgraph "Database"
            PostgreSQL[(PostgreSQL Database)]
            subgraph "Tables"
                Users[users]
                Characters[characters]
                Lists[lists]
                ListsUsers[lists_users]
                ListsSoulcores[lists_soulcores]
                CharactersSoulcores[characters_soulcores]
                ChatMessages[list_chat_messages]
                Creatures[creatures]
            end
        end
        
        Migrations[Goose Migrations<br/>- Schema Versioning<br/>- Database Evolution]
    end

    %% External Services
    subgraph "External Services"
        TibiaData[TibiaData API<br/>- Character Verification<br/>- World Information]
        EmailProvider[Email Service<br/>- Notifications<br/>- User Communications]
    end

    %% Infrastructure
    subgraph "Infrastructure"
        Docker[Docker Compose<br/>- Development Environment<br/>- Production Deployment]
        FileSystem[File System<br/>- Static Assets<br/>- Uploads]
    end

    %% Connections
    UI --> Views
    UI --> Components
    Views --> Stores
    Components --> Stores
    Stores --> APIServices
    APIServices --> Routes
    
    Routes --> Handlers
    Handlers --> Auth
    Handlers --> Services
    Handlers --> Validation
    Handlers --> SQLC
    Handlers --> AppError
    
    Services --> TibiaData
    Services --> EmailProvider
    
    SQLC --> Queries
    SQLC --> PostgreSQL
    Queries --> PostgreSQL
    Migrations --> PostgreSQL
    
    PostgreSQL --> Users
    PostgreSQL --> Characters
    PostgreSQL --> Lists
    PostgreSQL --> ListsUsers
    PostgreSQL --> ListsSoulcores
    PostgreSQL --> CharactersSoulcores
    PostgreSQL --> ChatMessages
    PostgreSQL --> Creatures
    
    Docker --> UI
    Docker --> Routes
    Docker --> PostgreSQL

    %% Styling
    classDef frontend fill:#e1f5fe
    classDef backend fill:#f3e5f5
    classDef database fill:#e8f5e8
    classDef external fill:#fff3e0
    classDef infrastructure fill:#fafafa
    
    class UI,Views,Components,Stores,APIServices,Router,I18n frontend
    class Routes,Handlers,Auth,Services,Validation,AppError backend
    class SQLC,Queries,PostgreSQL,Users,Characters,Lists,ListsUsers,ListsSoulcores,CharactersSoulcores,ChatMessages,Creatures,Migrations database
    class TibiaData,EmailProvider external
    class Docker,FileSystem infrastructure
```

## Key Architecture Components

### Frontend Architecture
- **Vue 3 with Composition API**: Modern reactive framework
- **TypeScript**: Type safety throughout the frontend
- **Pinia**: Centralized state management
- **TailwindCSS**: Utility-first styling
- **Vue I18n**: Internationalization support

### Backend Architecture
- **Echo Framework**: Fast HTTP router and middleware
- **SQLC**: Type-safe database queries from SQL
- **Custom Error System**: Structured error handling with context
- **JWT Authentication**: Secure user sessions
- **OAuth2**: Third-party authentication integration

### Database Design
- **PostgreSQL**: Robust relational database
- **Goose Migrations**: Version-controlled schema changes
- **Normalized Schema**: Efficient data relationships
- **UUID Primary Keys**: Globally unique identifiers

### Key Data Relationships
- Users can have multiple Characters
- Characters can be members of multiple Lists
- Lists contain Soul Cores with tracking states
- Characters can own Soul Cores (obtained/unlocked states)
- Lists have Chat functionality for member communication

### External Integrations
- **TibiaData API**: Character verification and game data
- **Email Services**: User notifications and communications
- **Docker**: Containerized development and deployment

### Testing Strategy
- **Backend**: Unit tests with gomock for database mocking
- **Frontend**: Component and integration testing
- **Type Safety**: SQLC and TypeScript provide compile-time safety
