# ![TibiaCores Logo](/frontend/public/favicon-32x32.png) TibiaCores

[![codecov](https://codecov.io/gh/sergot/tibiacores/graph/badge.svg?token=F29NYOLG42)](https://codecov.io/gh/sergot/tibiacores)

A web application for tracking and managing Tibia soulcore collections.

[TibiaCores.com](https://tibiacores.com)

## About

TibiaCores helps Tibia players track, manage, and share their soulcore collections. Create custom lists and collaborate with friends to efficiently complete your collection.


## Development

### Prerequisites

- Docker compose
- Go
- Node.js

### Development

1. Clone the repository:
   ```
   git clone https://github.com/sergot/tibiacores.git
   cd tibiacores
   ```

2. Start the application using Docker Compose:
   ```
   docker-compose up -d
   ```

3. Copy .env.example file to .env and set your environment variables:
   ```
   cp .env.example .env
   ```

4. Access the application at `http://localhost:5173`


## License

This project is licensed under the terms of the license included in the [LICENSE](LICENSE) file.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.