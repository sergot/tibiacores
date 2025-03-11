package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fiendlist/backend/handlers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	db          *mongo.Database
)

func initMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if we're in production environment
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	var mongoURI string
	if isProduction {
		// Use MongoDB Atlas in production
		mongoURI = os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			return fmt.Errorf("MONGODB_URI environment variable is required in production")
		}
		log.Println("Connecting to MongoDB Atlas in production mode")
	} else {
		// Use local MongoDB in development
		mongoURI = os.Getenv("MONGODB_URI")
		if mongoURI == "" {
			mongoURI = "mongodb://localhost:27017/fiendlist"
		}
		log.Println("Connecting to local MongoDB in development mode")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the MongoDB server to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("Connected to MongoDB!")
	mongoClient = client
	db = client.Database("fiendlist")
	return nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Initialize MongoDB
	if err := initMongoDB(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Configure CORS
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	if isProduction {
		// In production, only allow requests from your Vercel frontend
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://fiendlist.vercel.app" // Default production frontend URL
		}
		
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{frontendURL},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}))
		
		log.Printf("CORS configured for production, allowing origin: %s", frontendURL)
	} else {
		// In development, allow all origins
		e.Use(middleware.CORS())
		log.Println("CORS configured for development, allowing all origins")
	}

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Fiendlist API is running!")
	})

	// Setup API routes
	setupRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func setupRoutes(e *echo.Echo) {
	// Create handlers
	playerHandler := handlers.NewPlayerHandler(db)
	creatureHandler := handlers.NewCreatureHandler(db)
	listHandler := handlers.NewListHandler(db)

	// API group
	api := e.Group("/api")

	// Health check
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Player routes
	api.POST("/players", playerHandler.CreatePlayer)
	api.GET("/players/session/:sessionID", playerHandler.GetPlayerBySessionID)
	api.POST("/players/:playerID/characters", playerHandler.AddCharacter)
	api.GET("/players/:playerID/characters", playerHandler.GetCharacters)
	api.PUT("/players/:playerID/characters/:characterID/main", playerHandler.SetMainCharacter)
	api.POST("/players/convert", playerHandler.ConvertAnonymousToAccount)
	api.PUT("/players/:playerID/username", playerHandler.UpdateUsername)

	// Creature routes
	api.GET("/creatures", creatureHandler.GetAllCreatures)
	api.GET("/creatures/:endpoint", creatureHandler.GetCreatureByEndpoint)
	api.POST("/creatures/import", creatureHandler.ImportCreatures)

	// List routes
	api.POST("/lists", listHandler.CreateList)
	api.GET("/lists", listHandler.GetLists)
	api.GET("/lists/:id", listHandler.GetListByID)
	api.GET("/lists/share/:shareCode", listHandler.GetListByShareCode)
	api.POST("/lists/join", listHandler.JoinList)
	api.POST("/lists/:listID/soul-cores", listHandler.AddSoulCoreToList)
	api.PUT("/lists/:listID/soul-cores/:soulCoreID", listHandler.UpdateSoulCoreInList)
	api.GET("/characters/:characterID/lists", listHandler.GetListsByCharacterID)
}
