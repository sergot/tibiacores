package main

import (
	"context"
	"log"
	"os"

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

	// Public list endpoints
	api.GET("/lists/preview/:share_code", listsHandler.GetListPreview)
	api.POST("/lists/join/:share_code", listsHandler.JoinList)

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
	protected.POST("/lists", listsHandler.CreateList)
	protected.GET("/lists/:id", listsHandler.GetList)
	protected.POST("/lists/:id/soulcores", listsHandler.AddSoulcore)
	protected.PUT("/lists/:id/soulcores", listsHandler.UpdateSoulcoreStatus)

	protected.GET("/users/:user_id/characters", usersHandler.GetCharactersByUserId)
	protected.GET("/users/:user_id/lists", usersHandler.GetUserLists)
}

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Initialize OAuth providers
	auth.PrepareOAuthProviders()

	dbUrl := os.Getenv("DB_URL")
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

	// Recovery middleware
	// e.Use(middleware.Recover())

	// Body limit middleware
	// e.Use(middleware.BodyLimit("2M"))

	// Secure middleware
	// e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
	// 	XSSProtection:         "1; mode=block",
	// 	ContentTypeNosniff:    "nosniff",
	// 	XFrameOptions:         "SAMEORIGIN",
	// 	HSTSMaxAge:            3600,
	// 	ContentSecurityPolicy: "default-src 'self'",
	// }))

	// Rate limiter middleware for anonymous users
	// e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
	// 	Skipper: func(c echo.Context) bool {
	// 		// Skip rate limiting for authenticated users
	// 		return c.Get("is_anonymous") == false
	// 	},
	// 	Store: middleware.NewRateLimiterMemoryStore(60), // 60 requests
	// 	IdentifierExtractor: func(ctx echo.Context) (string, error) {
	// 		return ctx.RealIP(), nil
	// 	},
	// 	ErrorHandler: func(context echo.Context, err error) error {
	// 		return echo.NewHTTPError(429, "Too many requests")
	// 	},
	// 	DenyHandler: func(context echo.Context, identifier string, err error) error {
	// 		return echo.NewHTTPError(429, "Too many requests")
	// 	},
	// }))

	// CORS middleware
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins:  []string{"http://localhost:5173"}, // Frontend dev server
	// 	AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	// 	AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	// 	ExposeHeaders: []string{"X-Auth-Token"}, // Expose the auth token header
	// }))

	setupRoutes(e, connPool)
	e.Logger.Fatal(e.Start(":8080"))
}
