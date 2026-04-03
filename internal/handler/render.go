package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

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
	"initial": func(s string) string {
		r := []rune(s)
		if len(r) == 0 {
			return ""
		}
		return string(r[0])
	},
	// fmtMoney formats cents as "$1,234.56" or "-$1,234.56"
	"fmtMoney": func(cents int64) string {
		neg := cents < 0
		if neg {
			cents = -cents
		}
		dollars := cents / 100
		frac := cents % 100
		s := strconv.FormatInt(dollars, 10)
		var b strings.Builder
		n := len(s)
		for i, c := range s {
			if i > 0 && (n-i)%3 == 0 {
				b.WriteRune(',')
			}
			b.WriteRune(c)
		}
		result := fmt.Sprintf("$%s.%02d", b.String(), frac)
		if neg {
			return "-" + result
		}
		return result
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
