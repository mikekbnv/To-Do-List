package custom_middleware

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mikekbnv/To-Do-List/helper"
)

const Token_Cookie_Name = "token"
const Refresh_Token_Cookie_Name = "refresh_token"

// Authz validates token and authorizes users
func Authentication() echo.MiddlewareFunc {
	return auth()
}

func auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			clientToken, err := c.Request().Cookie(Token_Cookie_Name)
			if err != nil || clientToken.Value == "" {
				return c.Redirect(http.StatusFound, "/login")
			}
			clientRefreshToken, err := c.Request().Cookie(Refresh_Token_Cookie_Name)
			if err != nil || clientRefreshToken.Value == "" {
				return c.Redirect(http.StatusFound, "/login")
			}
			//log.Println("TOKEN:", clientToken)
			user, msg := helper.ValidateToken(clientToken.Value, clientRefreshToken.Value)
			if msg != "" {
				log.Println("\nMSG: ", msg)
				return c.Redirect(http.StatusFound, "/login")
			}
			//log.Println("VALID:", claims)
			cookiestoken := &http.Cookie{
				Name:  Token_Cookie_Name,
				Value: *user.Token,
			}
			cookiesrefresh := &http.Cookie{
				Name:  Refresh_Token_Cookie_Name,
				Value: *user.Refresh_token,
			}
			c.SetCookie(cookiestoken)
			c.SetCookie(cookiesrefresh)
			c.Set("email", *user.Email)
			c.Set("first_name", *user.First_name)
			c.Set("last_name", *user.Last_name)
			c.Set("uid", user.User_id)
			next(c)
			return
		}
	}
}

// func CkeckRoute() echo.MiddlewareFunc {
// 	return checkrout()
// }

// func checkrout() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) (err error) {
// 			tmp := c.Response().Status
// 			// rout := c.Response().
// 			log.Println(tmp)
// 			next(c)
// 			return
// 		}
// 	}
// }
