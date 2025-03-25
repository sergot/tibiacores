package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sergot/fiendlist/backend/auth"
	"github.com/sergot/fiendlist/backend/handlers"
)

func setupRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	api := e.Group("/api")

	// Public endpoints (no auth required)
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Handlers initialization
	usersHandler := handlers.NewUsersHandler(connPool)
	listsHandler := handlers.NewListsHandler(connPool)
	creaturesHandler := handlers.NewCreaturesHandler(connPool)
	oauthHandler := handlers.NewOAuthHandler(connPool)
	claimsHandler := handlers.NewClaimsHandler(connPool)

	// Start background claim checker
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Changed from Hour to Minute
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := claimsHandler.ProcessPendingClaims(); err != nil {
					log.Printf("Error processing pending claims: %v", err)
				}
			}
		}
	}()

	// Public list endpoints that allow optional auth
	optionalAuth := api.Group("", auth.OptionalAuthMiddleware)
	optionalAuth.GET("/lists/preview/:share_code", listsHandler.GetListPreview)
	optionalAuth.POST("/lists/join/:share_code", listsHandler.JoinList)
	optionalAuth.POST("/lists", listsHandler.CreateList)

	// User management routes
	api.POST("/signup", usersHandler.Signup)
	api.POST("/login", usersHandler.Login)

	// OAuth routes
	authGroup := api.Group("/auth")
	authGroup.GET("/oauth/:provider", oauthHandler.Login)
	authGroup.GET("/oauth/:provider/callback", oauthHandler.Callback)

	// Protected routes with auth middleware
	protected := api.Group("", auth.AuthMiddleware)
	protected.GET("/creatures", creaturesHandler.GetCreatures)
	protected.GET("/lists/:id", listsHandler.GetList)
	protected.POST("/lists/:id/soulcores", listsHandler.AddSoulcore)
	protected.PUT("/lists/:id/soulcores", listsHandler.UpdateSoulcoreStatus)
	protected.DELETE("/lists/:id/soulcores/:creature_id", listsHandler.RemoveSoulcore)

	// User endpoints
	protected.GET("/users/:user_id/characters", usersHandler.GetCharactersByUserId)
	protected.GET("/users/:user_id/lists", usersHandler.GetUserLists)
	protected.GET("/pending-suggestions", listsHandler.GetPendingSuggestions)

	// Character and suggestion endpoints
	protected.GET("/characters/:id", usersHandler.GetCharacter)
	protected.GET("/characters/:id/soulcores", usersHandler.GetCharacterSoulcores)
	protected.DELETE("/characters/:id/soulcores/:creature_id", usersHandler.RemoveCharacterSoulcore)
	protected.GET("/characters/:id/suggestions", listsHandler.GetCharacterSuggestions)
	protected.POST("/characters/:id/suggestions/accept", listsHandler.AcceptSoulcoreSuggestion)
	protected.POST("/characters/:id/suggestions/dismiss", listsHandler.DismissSoulcoreSuggestion)

	protected.POST("/claims", claimsHandler.StartClaim)
	protected.GET("/claims/:id", claimsHandler.CheckClaim)

}

func main() {
	ctx := context.Background()

	// Load .env file if it exists, ignore error in production
	_ = godotenv.Load()

	// Initialize OAuth providers
	auth.PrepareOAuthProviders()

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer connPool.Close()

	e := echo.New()

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Logger middleware with request ID
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${id} ${remote_ip} ${method} ${uri} ${status} ${latency_human}\n",
	}))

	setupRoutes(e, connPool)
	e.Logger.Fatal(e.Start(":8080"))
}
