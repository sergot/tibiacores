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
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	customMiddleware "github.com/sergot/tibiacores/backend/middleware"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/sergot/tibiacores/backend/services"
)

func setupRoutes(e *echo.Echo, emailService *services.EmailService, store db.Store) {
	api := e.Group("/api")

	// Public endpoints (no auth required)
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Handlers initialization
	usersHandler := handlers.NewUsersHandler(store, emailService)
	listsHandler := handlers.NewListsHandler(store)
	oauthHandler := handlers.NewOAuthHandler(store)
	claimsHandler := handlers.NewClaimsHandler(store)
	creaturesHandler := handlers.NewCreaturesHandler(store)
	charactersHandler := handlers.NewCharactersHandler(store)
	authHandler := handlers.NewAuthHandler(store)

	// Public endpoints
	api.GET("/creatures", creaturesHandler.GetCreatures)
	api.GET("/characters/public/:name", usersHandler.GetCharacterPublic)
	api.GET("/highscores", charactersHandler.GetHighscores)

	// Start background claim checker
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := claimsHandler.ProcessPendingClaims(); err != nil {
				log.Printf("Error processing pending claims: %v", err)
			}
		}
	}()

	// Public list endpoints that allow optional auth
	optionalAuth := api.Group("", auth.OptionalAuthMiddleware)
	optionalAuth.GET("/lists/preview/:share_code", listsHandler.GetListPreview)
	optionalAuth.POST("/lists/join/:share_code", listsHandler.JoinList)
	optionalAuth.POST("/lists", listsHandler.CreateList)

	// User management routes with rate limiting
	authGroup := api.Group("/auth", customMiddleware.RateLimitAuth())
	authGroup.POST("/signup", usersHandler.Signup)
	authGroup.POST("/login", usersHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout)
	authGroup.GET("/verify-email", usersHandler.VerifyEmail)

	// OAuth routes (using the same auth group with rate limiting)
	authGroup.GET("/oauth/:provider", oauthHandler.Login)
	authGroup.GET("/oauth/:provider/callback", oauthHandler.Callback)

	// Protected routes with auth middleware
	protected := api.Group("", auth.AuthMiddleware)
	protected.GET("/lists/:id", listsHandler.GetList)
	protected.GET("/lists/:id/members", listsHandler.GetListMembersWithUnlocks)
	protected.POST("/lists/:id/soulcores", listsHandler.AddSoulcore)
	protected.PUT("/lists/:id/soulcores", listsHandler.UpdateSoulcoreStatus)
	protected.DELETE("/lists/:id/soulcores/:creature_id", listsHandler.RemoveSoulcore)

	// User endpoints
	protected.GET("/users/:user_id/characters", usersHandler.GetCharactersByUserId)
	protected.GET("/users/:user_id/lists", usersHandler.GetUserLists)
	protected.GET("/users/:user_id", usersHandler.GetUser)
	protected.GET("/pending-suggestions", usersHandler.GetPendingSuggestions)

	// Character and suggestion endpoints
	protected.GET("/characters/:id", usersHandler.GetCharacter)
	protected.GET("/characters/:id/soulcores", usersHandler.GetCharacterSoulcores)
	protected.POST("/characters/:id/soulcores", usersHandler.AddCharacterSoulcore)
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
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found: %v", err)
		}
	}

	// Initialize OAuth providers
	auth.PrepareOAuthProviders()

	// Required environment variables
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		if os.Getenv("APP_ENV") == "production" {
			log.Fatal("FRONTEND_URL environment variable is required in production")
		}
		frontendURL = "http://localhost:5173" // Default for development
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && os.Getenv("APP_ENV") == "production" {
		log.Fatal("JWT_SECRET environment variable is required in production")
	}

	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	if refreshTokenSecret == "" && os.Getenv("APP_ENV") == "production" {
		log.Fatal("REFRESH_TOKEN_SECRET environment variable is required in production")
	}

	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer connPool.Close()

	e := echo.New()

	// Security headers middleware (applied globally)
	e.Use(customMiddleware.SecurityHeaders())

	// CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{frontendURL},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowCredentials: true, // Required for cookies to work in cross-origin requests
	}))

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Logger middleware with request ID
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${id} ${remote_ip} ${method} ${uri} ${status} ${latency_human}\n",
	}))

	// Custom error handling middleware
	e.Use(customMiddleware.RecoverWithConfig())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		// Use our custom error response handler
		httpErr := apperror.ErrorResponse(err)
		_ = c.JSON(httpErr.Code, httpErr.Message)
	}

	emailService, err := services.NewEmailService()
	if err != nil {
		log.Fatal("Error initializing email service: ", err)
	}

	store := db.NewStore(connPool)

	setupRoutes(e, emailService, store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
