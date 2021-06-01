package routing

import (
	"github.com/labstack/echo"
)

// Route ...
type Route struct {
	Method     string
	Path       string
	Handler    func(c echo.Context) error
	Middleware []echo.MiddlewareFunc
	Routes     []Route
}

// Register ...
func Register(echo *echo.Echo, routes []Route, prefix string) {
	for _, route := range routes {
		if len(route.Routes) > 0 {
			echo.Group(route.Path, route.Middleware...)
			Register(echo, route.Routes, route.Path)
		} else {
			path := prefix + route.Path
			echo.Add(route.Method, path, route.Handler, route.Middleware...)
		}
	}
}
