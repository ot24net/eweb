package eweb

import (
	"io"
	"sync"
	"text/template"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

type Template struct {
	*template.Template
}

func NewTemplate(tpl *template.Template) *Template {
	return &Template{tpl}
}

// Implements Renderer interface
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.ExecuteTemplate(w, name, data)
}

var (
	// Global instance
	defaultE     *Eweb
	defaultELock = sync.Mutex{}
)

func DebugMode() bool {
	return defaultE.Debug
}

// Struct for rendering map
type H map[string]interface{}

type Eweb struct {
	*echo.Echo
}

// Using global instance to manager router packages
func Default() *Eweb {
	defaultELock.Lock()
	defer defaultELock.Unlock()
	if defaultE == nil {
		defaultE = &Eweb{
			Echo: echo.New(),
		}
		defaultE.Debug = false
		// monitor middleware
		defaultE.Use(defaultE.Monitor)
	}
	return defaultE
}

func (e *Eweb) colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return color.Green(code)
	case code >= 300 && code < 400:
		return color.White(code)
	case code >= 400 && code < 500:
		return color.Yellow(code)
	default:
		return color.Red(code)
	}
}

func (e *Eweb) colorForMethod(method string) string {
	switch method {
	case "GET":
		return color.Blue(method)
	case "POST":
		return color.Cyan(method)
	case "PUT":
		return color.Yellow(method)
	case "DELETE":
		return color.Red(method)
	case "PATCH":
		return color.Green(method)
	case "HEAD":
		return color.Magenta(method)
	case "OPTIONS":
		return color.White(method)
	default:
		return color.Reset(method)
	}
}

func (e *Eweb) Monitor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// request log
		defer func(start time.Time) {
			stop := time.Now()
			req := c.Request()
			res := c.Response()

			n := res.Status
			if !e.Debug && n < 400 {
				return
			}

			contentInL := req.Header.Get(echo.HeaderContentLength)
			if contentInL == "" {
				contentInL = "0"
			}
			contentOutL := res.Size
			color.Printf(
				"[echo] %s | %s | %s | %s | %s | %s | %sB | %dB \n",
				start.Format("2006-01-02 15:04:05"),
				e.colorForStatus(n), e.colorForMethod(req.Method), req.RequestURI,
				stop.Sub(start).String(), // latency_human
				c.RealIP(),
				contentInL, contentOutL,
			)
		}(time.Now())
		return next(c)
	}
}
