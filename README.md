# ![TibiaCores Logo](/frontend/public/favicon-32x32.png) TibiaCores

[![codecov](https://codecov.io/gh/sergot/tibiacores/graph/badge.svg?token=F29NYOLG42)](https://codecov.io/gh/sergot/tibiacores)

A web application for tracking and managing Tibia soulcore collections.

[TibiaCores.com](https://tibiacores.com)

## Features

- **Soul Core Tracking**: Track which Soul Cores you've obtained and unlocked for your characters.
- **Collaborative Lists**: Create detailed lists of Soul Cores and invite friends to collaborate. Perfect for hunting teams or guilds.
- **Group Chat**: Each collaborative list includes a real-time chat for coordination, complete with unread message indicators.
- **Character Verification**: Securely verify your Tibia characters using a challenge-response verification system powered by the TibiaData API.
- **Smart Suggestions**: Receive contextual suggestions for which Soul Cores to hunt next based on your current progress.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/sergot/tibiacores.git
cd tibiacores

# Copy environment configuration
cp .env.example .env

# Start with Docker Compose
docker compose up -d

# Access at http://localhost:5173
```

**For detailed setup instructions**, including OAuth configuration, email services, and manual setup without Docker, see [docs/setup.md](docs/setup.md).

## Technology Stack

- **Backend**: Go 1.25+, Echo Framework, PostgreSQL 17
- **Frontend**: Vue 3, TypeScript, TailwindCSS, Pinia
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Authentication**: JWT, OAuth2 (Discord, Google)
- **External Services**: Mailgun (EU), EmailOctopus, TibiaData API

## Documentation

- [Setup Guide](docs/setup.md) - Development environment setup
- [API Reference](docs/api-reference.md) - API endpoints documentation
- [Database Schema](docs/database.md) - Database structure and migrations
- [Architecture](docs/architecture.md) - System architecture and design
- [Contributing](CONTRIBUTING.md) - Contribution guidelines


## License

This project is licensed under the terms of the license included in the [LICENSE](LICENSE) file.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.