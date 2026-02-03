package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"
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
	"github.com/sergot/tibiacores/backend/pkg/validator"
	"github.com/sergot/tibiacores/backend/services"
)

func setupRoutes(e *echo.Echo, emailService *services.EmailService, newsletterService *services.NewsletterService, store db.Store, logger *slog.Logger) {
	// Create rate limiters
	globalLimiter := customMiddleware.NewIPRateLimiter(20, 40)  // 20 req/sec, burst of 40
	authLimiter := customMiddleware.NewIPRateLimiter(5.0/60, 5) // 5 req/min, burst of 5

	api := e.Group("/api")

	// Apply global rate limiting to all API routes
	api.Use(customMiddleware.RateLimiterMiddleware(globalLimiter))

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
	newsletterHandler := handlers.NewNewsletterHandler(newsletterService)

	// Public endpoints
	api.GET("/creatures", creaturesHandler.GetCreatures)
	api.GET("/characters/public/:name", usersHandler.GetCharacterPublic)
	api.GET("/highscores", charactersHandler.GetHighscores)
	api.POST("/newsletter/subscribe", newsletterHandler.Subscribe, customMiddleware.RateLimiterMiddleware(authLimiter))

	// Start background claim checker with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("claim processor panicked",
					"panic", r,
					"stack", string(debug.Stack()),
				)
			}
		}()

		// TODO: Add Distributed Lock (e.g. Postgres Advisory Lock) here for horizontal scaling support
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			logger.Info("starting claim check cycle")
			if err := claimsHandler.ProcessPendingClaims(); err != nil {
				logger.Error("error processing pending claims", "error", err)
			}
		}
	}()

	// Public list endpoints that allow optional auth
	optionalAuth := api.Group("", auth.OptionalAuthMiddleware)
	optionalAuth.GET("/lists/preview/:share_code", listsHandler.GetListPreview)
	optionalAuth.POST("/lists/join/:share_code", listsHandler.JoinList)
	optionalAuth.POST("/lists", listsHandler.CreateList)

	// User management routes (rate limited)
	api.POST("/signup", usersHandler.Signup, customMiddleware.RateLimiterMiddleware(authLimiter))
	api.POST("/login", usersHandler.Login, customMiddleware.RateLimiterMiddleware(authLimiter))
	api.GET("/verify-email", usersHandler.VerifyEmail)

	// OAuth routes
	authGroup := api.Group("/auth")
	authGroup.GET("/oauth/:provider", oauthHandler.Login)
	authGroup.GET("/oauth/:provider/callback", oauthHandler.Callback)

	// Protected routes with auth middleware
	protected := api.Group("", auth.AuthMiddleware)
	protected.GET("/lists/:id", listsHandler.GetList)
	protected.GET("/lists/:id/members", listsHandler.GetListMembersWithUnlocks)
	protected.POST("/lists/:id/soulcores", listsHandler.AddSoulcore)
	protected.PUT("/lists/:id/soulcores", listsHandler.UpdateSoulcoreStatus)
	protected.DELETE("/lists/:id/soulcores/:creature_id", listsHandler.RemoveSoulcore)

	// Chat endpoints
	protected.POST("/lists/:id/chat/read", listsHandler.MarkChatMessagesAsRead)
	protected.GET("/lists/:id/chat", listsHandler.GetChatMessages)
	protected.POST("/lists/:id/chat", listsHandler.CreateChatMessage)
	protected.DELETE("/lists/:id/chat/:messageId", listsHandler.DeleteChatMessage)
	protected.GET("/chat-notifications", listsHandler.GetChatNotifications)

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
	protected.GET("/claims/:id", claimsHandler.CheckClaim, customMiddleware.RateLimiterMiddleware(authLimiter))
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// Load .env file if it exists, ignore error in production
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			logger.Warn("Warning: .env file not found", "error", err)
		}
	}

	// Initialize OAuth providers
	auth.PrepareOAuthProviders()

	// Required environment variables
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		logger.Error("DB_URL environment variable is required")
		os.Exit(1)
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		if os.Getenv("APP_ENV") == "production" {
			logger.Error("FRONTEND_URL environment variable is required in production")
			os.Exit(1)
		}
		frontendURL = "http://localhost:5173" // Default for development
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && os.Getenv("APP_ENV") == "production" {
		logger.Error("JWT_SECRET environment variable is required in production")
		os.Exit(1)
	}

	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		logger.Error("Error connecting to the database", "error", err)
		os.Exit(1)
	}
	defer connPool.Close()

	e := echo.New()

	// Register validator
	e.Validator = validator.New()

	// Modern Security & Observability Middleware
	e.Use(middleware.Recover())
	e.Use(customMiddleware.SlogLogger(logger)) // Use custom slog logger
	e.Use(middleware.RequestID())

	// Security: CORS to allow frontend access
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "https://tibiacores.com", frontendURL},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Request-ID"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Auth-Token"},
	}))

	// Security: Limit body size to prevent DoS (2MB limit)
	e.Use(middleware.BodyLimit("2M"))

	// Security: Add secure headers
	e.Use(middleware.Secure())

	// Custom error handling middleware
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		// Use our custom error response handler
		httpErr := apperror.ErrorResponse(err)
		_ = c.JSON(httpErr.Code, httpErr.Message)
	}

	emailService, err := services.NewEmailService()
	if err != nil {
		logger.Error("Error initializing email service", "error", err)
		os.Exit(1)
	}

	newsletterService, err := services.NewNewsletterService()
	if err != nil {
		logger.Error("Error initializing newsletter service", "error", err)
		os.Exit(1)
	}

	store := db.NewStore(connPool)

	setupRoutes(e, emailService, newsletterService, store, logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Start server with error logging
	if err := e.Start(":" + port); err != nil {
		logger.Error("Server shutdown", "error", err)
		os.Exit(1)
	}
}
