package chi

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/savioxavier/termlink"
)

type Route struct {
	route.Router
	config             config.Config
	htmlRender         *template.Template
	instance           *chi.Mux
	maxMultipartMemory int64
	server             *http.Server
	tlsServer          *http.Server
}

func NewRoute(config config.Config, parameters map[string]any) (*Route, error) {
	mux := chi.NewRouter()
	if debugLog := getDebugLog(config); debugLog != nil {
		mux.Use(debugLog)
	}

	htmlRender := new(template.Template)
	if driver, exist := parameters["driver"]; exist {
		newHtmlRender, ok := config.Get("http.drivers." + driver.(string) + ".template").(*template.Template)
		if ok {
			htmlRender = newHtmlRender
		} else {
			htmlRenderCallback, ok := config.Get("http.drivers." + driver.(string) + ".template").(func() (*template.Template, error))
			if ok {
				newHtmlRender, err := htmlRenderCallback()
				if err != nil {
					return nil, err
				}

				htmlRender = newHtmlRender
			}
		}
	}

	return &Route{
		Router: NewGroup(
			config,
			mux,
			"",
			[]httpcontract.Middleware{},
			[]httpcontract.Middleware{ResponseMiddleware()},
		),
		config:             config,
		htmlRender:         htmlRender,
		instance:           mux,
		maxMultipartMemory: int64(config.GetInt("http.drivers.chi.body_limit", 4096)) << 10,
	}, nil
}

func (r *Route) Fallback(handler httpcontract.HandlerFunc) {
	r.instance.NotFound(handlerToChiHandler(handler))
}

func (r *Route) GlobalMiddleware(middlewares ...httpcontract.Middleware) {
	middlewares = append(middlewares, Cors(), Tls())
	r.instance.Use(middlewaresToChiHandlers(middlewares)...)
	r.Router = NewGroup(
		r.config,
		r.instance,
		"",
		[]httpcontract.Middleware{},
		[]httpcontract.Middleware{ResponseMiddleware()},
	)
}

func (r *Route) Run(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.host")
		defaultPort := r.config.GetString("http.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	r.outputRoutes()
	color.Green().Println(termlink.Link("[HTTP] Listening and serving HTTP on", "http://"+host[0]))

	r.server = &http.Server{
		Addr:           host[0],
		Handler:        http.AllowQuerySemicolons(r.instance),
		MaxHeaderBytes: r.config.GetInt("http.drivers.chi.header_limit", 4096) << 10,
	}

	if err := r.server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

func (r *Route) RunTLS(host ...string) error {
	if len(host) == 0 {
		defaultHost := r.config.GetString("http.tls.host")
		defaultPort := r.config.GetString("http.tls.port")
		if defaultPort == "" {
			return errors.New("port can't be empty")
		}
		completeHost := defaultHost + ":" + defaultPort
		host = append(host, completeHost)
	}

	certFile := r.config.GetString("http.tls.ssl.cert")
	keyFile := r.config.GetString("http.tls.ssl.key")

	return r.RunTLSWithCert(host[0], certFile, keyFile)
}

func (r *Route) RunTLSWithCert(host, certFile, keyFile string) error {
	if host == "" {
		return errors.New("host can't be empty")
	}
	if certFile == "" || keyFile == "" {
		return errors.New("certificate can't be empty")
	}

	r.outputRoutes()
	color.Green().Println(termlink.Link("[HTTPS] Listening and serving HTTPS on", "https://"+host))

	r.tlsServer = &http.Server{
		Addr:           host,
		Handler:        http.AllowQuerySemicolons(r.instance),
		MaxHeaderBytes: r.config.GetInt("http.drivers.chi.header_limit", 4096) << 10,
	}

	if err := r.tlsServer.ListenAndServeTLS(certFile, keyFile); errors.Is(err, http.ErrServerClosed) {
		return nil
	} else {
		return err
	}
}

func (r *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.instance.ServeHTTP(writer, request)
}

func (r *Route) Shutdown(ctx ...context.Context) error {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}

	if r.server != nil {
		return r.server.Shutdown(c)
	}
	if r.tlsServer != nil {
		return r.tlsServer.Shutdown(c)
	}
	return nil
}

func (r *Route) outputRoutes() {
	if r.config.GetBool("app.debug") && support.Env != support.EnvArtisan {
		walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			route = strings.Replace(route, "/*/", "/", -1)
			fmt.Printf("%-10s %s\n", method, route)
			return nil
		}

		_ = chi.Walk(r.instance, walkFunc)
	}
}
