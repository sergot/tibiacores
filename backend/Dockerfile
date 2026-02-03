FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# For development with hot reload
FROM golang:1.25-alpine AS dev

WORKDIR /app

# Install Air for hot reloading
RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Command to run Air for hot reloading
CMD ["air", "-c", ".air.toml"]

# Production image
FROM alpine:latest AS prod

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# copy migrations and install goose
COPY --from=builder /app/db/migrations ./migrations
RUN apk add --no-cache go
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]