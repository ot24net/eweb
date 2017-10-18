package eweb

import (
	"io"
	"net"
	"net/http"
	"os"
	"strings"
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
	defaultE     *Echo
	defaultELock = sync.Mutex{}
)

func DebugMode() bool {
	return defaultE.Debug
}

// Struct for rendering map
type H map[string]interface{}

type Echo struct {
	*echo.Echo
	colorer *color.Color
}

// Using global instance to manager router packages
func Default() *Echo {
	defaultELock.Lock()
	defer defaultELock.Unlock()
	if defaultE == nil {
		defaultE = &Echo{Echo: echo.New(), colorer: color.New()}
		defaultE.Debug = os.Getenv("GIN_MODE") != "release"
		defaultE.HideBanner = true
		defaultE.Echo.Server.Handler = defaultE
		defaultE.Echo.TLSServer.Handler = defaultE
	}
	return defaultE
}

func (e *Echo) colorForStatus(code string) string {
	switch {
	case code >= "200" && code < "300":
		return e.colorer.Green(code)
	case code >= "300" && code < "400":
		return e.colorer.White(code)
	case code >= "400" && code < "500":
		return e.colorer.Yellow(code)
	default:
		return e.colorer.Red(code)
	}
}

func (e *Echo) colorForMethod(method string) string {
	switch method {
	case "GET":
		return e.colorer.Blue(method)
	case "POST":
		return e.colorer.Cyan(method)
	case "PUT":
		return e.colorer.Yellow(method)
	case "DELETE":
		return e.colorer.Red(method)
	case "PATCH":
		return e.colorer.Green(method)
	case "HEAD":
		return e.colorer.Magenta(method)
	case "OPTIONS":
		return e.colorer.White(method)
	default:
		return e.colorer.Reset(method)
	}
}

func getRealIp(request *http.Request) string {
	ra := request.RemoteAddr
	if ip := request.Header.Get(echo.HeaderXForwardedFor); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := request.Header.Get(echo.HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra

}

// rebuild echo.Echo#ServeHTTP
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// request log
	defer func(start time.Time) {
		stop := time.Now()
		req := r

		n := w.Header().Get("STATUS")
		if !e.Debug && n < "400" {
			return
		}
		e.colorer.Printf(
			"[echo] %s | %s | %s | %s | %s | %s \n",
			start.Format("2006-01-02 15:04:05"),
			e.colorForStatus(n), e.colorForMethod(req.Method), req.RequestURI,
			stop.Sub(start).String(), // latency_human
			getRealIp(r),
		)

	}(time.Now())

	// super call
	e.Echo.ServeHTTP(w, r)

}
