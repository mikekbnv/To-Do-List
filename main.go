package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mikekbnv/To-Do-List/config"
	"github.com/mikekbnv/To-Do-List/internal/routing"
	"github.com/mikekbnv/To-Do-List/internal/template"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e := echo.New()
	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}),
		//middle.CkeckRoute(),
	)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.Renderer = template.NewEngine("web")
	e.Static("/", "web")
	routing.Register(e, config.Routes, "")

	e.Logger.Fatal(e.Start(":" + port))
}
