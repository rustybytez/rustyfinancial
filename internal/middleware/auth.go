package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := os.Getenv("AUTH_TOKEN")
		if token == "" {
			return next(c)
		}
		cookie, err := c.Cookie("auth_token")
		if err != nil || cookie.Value != token {
			path := c.Request().URL.Path
			if path == "/login" || path == "/health" || strings.HasPrefix(path, "/static/") {
				return next(c)
			}
			return c.Redirect(http.StatusFound, "/login")
		}
		return next(c)
	}
}
