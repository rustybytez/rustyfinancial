package handler

import (
	"html/template"
	"net/http"

	"rustyfinancial/web"

	"github.com/labstack/echo/v4"
)

var funcMap = template.FuncMap{
	"divf": func(a, b int64) float64 { return float64(a) / float64(b) },
	"list": func(vals ...string) []string { return vals },
	"dict": func(pairs ...string) map[string]string {
		m := make(map[string]string, len(pairs)/2)
		for i := 0; i+1 < len(pairs); i += 2 {
			m[pairs[i]] = pairs[i+1]
		}
		return m
	},
}

func render(c echo.Context, page string, data any) error {
	t, err := template.New("").Funcs(funcMap).ParseFS(web.TemplateFS, "templates/layout.html", "templates/"+page)
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return t.ExecuteTemplate(c.Response().Writer, "layout", data)
}
