package main

import (
	"log"
	"net/http"
	"os"

	"rustyfinancial/internal/handler"
	appmiddleware "rustyfinancial/internal/middleware"
	"rustyfinancial/internal/store"
	"rustyfinancial/web"

	"io/fs"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "rustyfinancial.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s, err := store.New(dsn)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(appmiddleware.Auth)

	// Static files
	staticFS, _ := fs.Sub(web.StaticFS, "static")
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))

	// Health
	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	// Auth
	e.GET("/login", handler.LoginPage)
	e.POST("/login", handler.Login)
	e.GET("/logout", handler.Logout)

	// Accounts
	ah := handler.NewAccountHandler(s)
	e.GET("/", ah.Dashboard)
	e.GET("/accounts/new", ah.NewForm)
	e.POST("/accounts", ah.Create)
	e.GET("/accounts/:id/edit", ah.EditForm)
	e.POST("/accounts/:id/edit", ah.Update)
	e.POST("/accounts/:id/delete", ah.Delete)

	// Transactions
	th := handler.NewTransactionHandler(s)
	e.GET("/accounts/:id", th.List)
	e.POST("/accounts/:id/transactions", th.Create)
	e.POST("/accounts/:id/transactions/:txid/delete", th.Delete)

	log.Printf("listening on :%s", port)
	log.Fatal(e.Start(":" + port))
}
