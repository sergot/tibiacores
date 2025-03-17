package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sergot/fiendlist/backend/db"
	"github.com/sergot/fiendlist/backend/handlers"
)

func setupRoutes(e *echo.Echo, queries *db.Queries) {
	api := e.Group("/api")

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	creaturesHandler := handlers.NewCreaturesHandler(queries)
	api.GET("/creatures", creaturesHandler.GetCreatures)
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dbUrl := os.Getenv("DB_URI")

	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	setupRoutes(e, queries)

	e.Logger.Fatal(e.Start(":8080"))
}
