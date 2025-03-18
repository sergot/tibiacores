package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sergot/fiendlist/backend/handlers"
)

func setupRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	api := e.Group("/api")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	creaturesHandler := handlers.NewCreaturesHandler(connPool)
	api.GET("/creatures", creaturesHandler.GetCreatures)

	listsHandler := handlers.NewListsHandler(connPool)
	api.GET("/lists/:user_id", listsHandler.GetUserLists)
	api.POST("/lists", listsHandler.CreateList)

	usersHandler := handlers.NewUsersHandler(connPool)
	api.GET("/users/:user_id/characters", usersHandler.GetCharactersByUserId)
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dbUrl := os.Getenv("DB_URL")

	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer connPool.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	setupRoutes(e, connPool)

	e.Logger.Fatal(e.Start(":8080"))
}
