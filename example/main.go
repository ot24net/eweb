package main

import (
	"log"
	"strings"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ot24net/eweb"
	_ "github.com/ot24net/eweb/example/routes"
)

// register static path
func init() {
	e := eweb.Default()
	e.Static("/", "./public")
}

func main() {
	e := eweb.Default()

	// render
	e.Renderer = eweb.NewTemplate(
		template.Must(template.ParseGlob("./public/**/tpl/*.html")),
	)

	// middle ware
	e.Use(middleware.Gzip())

	// filter
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			uri := req.URL.Path
			switch {
			case strings.HasPrefix(uri, "/hacheck"):
				// live check
				return c.String(200, "1")
			case uri == "/":
				// TODO: redirect to need
				// return c.Redirect(301,"/index")
			}
			return next(c)
		}
	})

	// Start server
	addr := ":8081"
	log.Printf("Listen: %s\n", addr)
	log.Fatal(e.Start(addr))
}
