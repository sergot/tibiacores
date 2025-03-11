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

## Deployment

This project is set up for deployment with:
- Frontend: Vercel
- Backend: Render
- Database: MongoDB Atlas

### Prerequisites

- GitHub account
- Vercel account
- Render account
- MongoDB Atlas account

### MongoDB Atlas Setup

1. Create a MongoDB Atlas account at [https://www.mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
2. Create a new cluster (free tier is sufficient to start)
3. Set up database access:
   - Create a database user with a strong password
   - Give this user read/write access to your database
4. Configure network access:
   - Initially, allow access from anywhere (0.0.0.0/0)
   - Later, restrict this to your backend's IP addresses
5. Get your connection string from the "Connect" button

### Backend Deployment on Render

1. Fork or clone this repository to your GitHub account
2. Sign up at [render.com](https://render.com)
3. Connect your GitHub repository
4. Create a new Web Service:
   - Select your repository
   - Name: `fiendlist-backend` (or your preferred name)
   - Environment: Go
   - Build Command: `cd backend && go build -o server`
   - Start Command: `cd backend && ./server`
   - Add the following environment variables:
     - `ENVIRONMENT`: `production`
     - `MONGODB_URI`: Your MongoDB Atlas connection string
     - `PORT`: `8080`
     - `FRONTEND_URL`: Your Vercel frontend URL (e.g., `https://fiendlist.vercel.app`)

### Frontend Deployment on Vercel

1. Sign up at [vercel.com](https://vercel.com)
2. Import your GitHub repository
3. Configure the project:
   - Framework Preset: Next.js
   - Root Directory: `frontend`
   - Build Command: `npm run build`
   - Output Directory: `.next`
   - Add the following environment variable:
     - `NEXT_PUBLIC_API_URL`: Your Render backend URL + `/api` (e.g., `https://fiendlist-backend.onrender.com/api`)
4. Deploy

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

### Production Deployment

1. Create a `.env` file with your production settings:
   ```bash
   echo "JWT_SECRET=your_secure_jwt_secret" > .env
   ```

2. Start the application in production mode:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080/api

### Backend Setup (Manual)

1. Navigate to the backend directory:
   ```
   cd backend
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Create a `.env` file with the following content:
   ```
   PORT=8080
   MONGODB_URI=mongodb://localhost:27017/fiendlist
   JWT_SECRET=your_jwt_secret_key_change_in_production
   ```

4. Run the backend:
   ```
   go run main.go
   ```

### Frontend Setup (Manual)

1. Navigate to the frontend directory:
   ```
   cd frontend
   ```

2. Install dependencies:
   ```
   npm install
   ```

3. Run the frontend:
   ```
   npm run dev
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

## API Endpoints

### Players
- `POST /api/players` - Create a new player
- `GET /api/players/session/:sessionID` - Get a player by session ID

### Creatures
- `GET /api/creatures` - Get all creatures
- `GET /api/creatures/:endpoint` - Get a creature by endpoint
- `POST /api/creatures/import` - Import creatures from JSON file

### Soulpit Lists
- `POST /api/soulpit-lists` - Create a new soulpit list
- `GET /api/soulpit-lists/:id` - Get a soulpit list by ID
- `GET /api/soulpit-lists/share/:shareCode` - Get a soulpit list by share code
- `GET /api/soulpit-lists/player/:playerID` - Get all soulpit lists for a player
- `POST /api/soulpit-lists/:listID/players` - Add a player to a soulpit list
- `POST /api/soulpit-lists/:listID/soul-cores` - Add a soul core to a soulpit list
- `PUT /api/soulpit-lists/:listID/soul-cores/:soulCoreID` - Update a soul core in a soulpit list

## Future Premium Features

- Advanced statistics and analytics
- Custom list themes and styling
- Priority notifications for rare Soul Cores
- Integration with Tibia.com API (if available)
- Export/import lists in various formats
- Team management features for guilds

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 