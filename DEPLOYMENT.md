# Deployment Guide

This guide explains how to deploy the Fiendlist application using Vercel for the frontend, Render for the backend, and MongoDB Atlas for the database.

## Prerequisites

- GitHub account
- Vercel account
- Render account
- MongoDB Atlas account

## MongoDB Atlas Setup

1. Create a MongoDB Atlas account at [https://www.mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
2. Create a new cluster (free tier is sufficient to start)
3. Set up database access:
   - Create a database user with a strong password
   - Give this user read/write access to your database
4. Configure network access:
   - Initially, allow access from anywhere (0.0.0.0/0)
   - Later, restrict this to your backend's IP addresses
5. Get your connection string from the "Connect" button

## Backend Deployment on Render

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

## Frontend Deployment on Vercel

1. Sign up at [vercel.com](https://vercel.com)
2. Import your GitHub repository
3. Configure the project:
   - Framework Preset: Next.js
   - Root Directory: `frontend`
   - Build Command: `npm run build` (or your custom build command)
   - Output Directory: `.next`
   - Add the following environment variable:
     - `NEXT_PUBLIC_API_URL`: Your Render backend URL + `/api` (e.g., `https://fiendlist-backend.onrender.com/api`)
4. Deploy

## Updating Your Deployment

### Backend Updates

1. Push changes to your GitHub repository
2. Render will automatically rebuild and deploy your backend

### Frontend Updates

1. Push changes to your GitHub repository
2. Vercel will automatically rebuild and deploy your frontend

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