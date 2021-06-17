package config

import (
	"github.com/labstack/echo"
	"github.com/mikekbnv/To-Do-List/internal/routers"
	"github.com/mikekbnv/To-Do-List/internal/routing"
	custom_middleware "github.com/mikekbnv/To-Do-List/middleware"
)

// Routes ...
var Routes = []routing.Route{
	{
		Method:  "GET",
		Path:    "/",
		Handler: routers.Signup_Form,
	},
	{
		Method:  "POST",
		Path:    "/",
		Handler: routers.Signup,
	},
	{
		Method:  "GET",
		Path:    "/login",
		Handler: routers.Login_Form,
	},
	{
		Method:  "POST",
		Path:    "/login",
		Handler: routers.Login,
	},
	{
		Method:  "POST",
		Path:    "/logout",
		Handler: routers.Logout,
	},
	{
		Method:  "GET",
		Path:    "/list",
		Handler: routers.Get_List,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/list",
		Handler: routers.Createtask,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/delete",
		Handler: routers.Deletetask,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "GET",
		Path:    "/alltasks",
		Handler: routers.All_tasks,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/undo",
		Handler: routers.Undo_task,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/edit",
		Handler: routers.Edit_task,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "GET",
		Path:    "/account",
		Handler: routers.Account_info,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/account",
		Handler: routers.Account_info_page,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
	{
		Method:  "POST",
		Path:    "/account/update",
		Handler: routers.Account_update,
		Middleware: []echo.MiddlewareFunc{
			custom_middleware.Authentication(),
		},
	},
}
