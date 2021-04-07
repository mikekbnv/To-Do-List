package main

import (
	"os"

	"github.com/mikekbnv/To-Do-List/internal/routers"
	"github.com/mikekbnv/To-Do-List/internal/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e := echo.New()

	e.Use(
		middleware.Logger(),
		middleware.Recover(),
	)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.Renderer = template.NewEngine("web")
	e.Static("/", "web")
	routers.Register(e)

	e.Logger.Fatal(e.Start(":" + port))
}
