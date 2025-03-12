# Fiendlist - Tibia Soul Pit Manager

Fiendlist is a web application designed to help Tibia players manage and organize their Soul Pit activities. It allows players to create lists of Soul Cores, invite other players, and track their progress towards unlocking experience boosts.

## Features

- Create Soul Pit lists and invite other players
- Add Soul Cores to lists and track which ones have been obtained
- Share lists with other players via a unique share code
- No registration required - just enter your in-game nickname
- Track progress towards experience boost milestones

## Technology Stack

### Backend
- Go with Echo web framework
- MongoDB for database
- JWT for authentication

### Frontend
- Vue.js 3 with Nuxt 3
- Nuxt UI for components
- Tailwind CSS for styling
- Pinia for state management


## Local Development

For local development, use the Docker Compose setup:

```bash
docker-compose up
```

This will start the MongoDB, backend, and frontend services locally.

## Environment Variables

### Backend

- `ENVIRONMENT`: Set to `development` for local, `production` for deployed
- `MONGODB_URI`: MongoDB connection string
- `PORT`: Server port (default: 8080)
- `FRONTEND_URL`: Frontend URL for CORS configuration

### Frontend

- `NEXT_PUBLIC_API_URL`: Backend API URL

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- MongoDB
- Docker and Docker Compose (for containerized setup)

### Using Docker Compose (Recommended)

1. Start all services:
   ```
   docker-compose up -d
   ```

2. Access the application at http://localhost:3000

3. To reset the database after schema changes:
   ```
   ./reset-db.sh
   ```


## Development

### Hot Reloading

Both the frontend and backend support hot reloading for a better development experience:

- **Frontend**: Uses Nuxt's built-in hot module replacement (HMR)
- **Backend**: Uses Air for live reloading

When you make changes to your code, the applications will automatically rebuild and refresh.

## Project Structure

```
fiendlist/
├── backend/             # Backend Go application with Echo framework
├── frontend/            # Frontend Nuxt.js application
│   ├── pages/           # Nuxt pages
│   ├── components/      # Vue components
│   ├── stores/          # Pinia stores
│   ├── services/        # API services
│   ├── public/          # Static assets
│   └── nuxt.config.ts   # Nuxt configuration
├── docker-compose.yml   # Docker Compose configuration for development
├── docker-compose.prod.yml # Docker Compose configuration for production
└── README.md            # This file
```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 