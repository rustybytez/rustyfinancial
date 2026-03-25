package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func LoginPage(c echo.Context) error {
	return render(c, "auth/login.html", nil)
}

func Login(c echo.Context) error {
	token := c.FormValue("token")
	if token != os.Getenv("AUTH_TOKEN") {
		return render(c, "auth/login.html", map[string]any{"Error": "Invalid token"})
	}
	c.SetCookie(&http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(90 * 24 * time.Hour),
	})
	return c.Redirect(http.StatusFound, "/")
}

func Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:    "auth_token",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	return c.Redirect(http.StatusFound, "/login")
}
