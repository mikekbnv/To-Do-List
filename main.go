package main

import (
	"To-Do-List/routers"
	"To-Do-List/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
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

	e.Logger.Fatal(e.Start(":8000"))
}
