package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
)

func run() error {
	// TODO: context

	e := echo.New()

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.POST("/api/creatures", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	return e.Start(":8080")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
