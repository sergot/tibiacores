#!/bin/bash

# Stop all containers
echo "Stopping all containers..."
docker-compose down

# Remove containers if they still exist
echo "Removing containers if they still exist..."
docker rm -f fiendlist-mongodb fiendlist-backend fiendlist-frontend fiendlist-init-db 2>/dev/null || true

# Remove the MongoDB volume
echo "Removing MongoDB volume..."
docker volume rm fiendlist_mongodb_data 2>/dev/null || true

# Start all containers again
echo "Starting all containers with a fresh database..."
docker-compose up -d

# Wait for MongoDB to initialize
echo "Waiting for MongoDB to initialize..."
sleep 5

# Check if creatures were imported
echo "Checking if creatures were imported..."
docker exec fiendlist-mongodb mongosh fiendlist --eval "db.creatures.countDocuments()"

# Check if backend is running
echo "Checking if backend is running..."
curl http://localhost:8080/api/health

echo "Database reset complete!" 