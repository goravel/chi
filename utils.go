package chi

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
)

func middlewaresToChiHandlers(instance *Instance, middlewares []httpcontract.Middleware) []func(http.Handler) http.Handler {
	var handlers []func(http.Handler) http.Handler
	for _, item := range middlewares {
		handlers = append(handlers, middlewareToChiHandler(instance, item))
	}

	return handlers
}

func handlerToChiHandler(instance *Instance, handler httpcontract.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if response := handler(NewContext(instance, w, r)); response != nil {
			_ = response.Render()
		}
	}
}

func middlewareToChiHandler(instance *Instance, handler httpcontract.Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler(NewContext(instance, w, r))
			next.ServeHTTP(w, r)
		})
	}
}

func getDebugLog(config config.Config) func(next http.Handler) http.Handler {
	if config.GetBool("app.debug") {
		return middleware.Logger
	}

	return nil
}

// TODO optimize this to avoid copying the request body every time
func copyRequest(r *http.Request) *http.Request {
	newRequest := r.Clone(r.Context())
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))
		newRequest.Body = io.NopCloser(bytes.NewReader(body))
	}

	return newRequest
}

func mergeSlashForPath(path string) string {
	path = strings.ReplaceAll(path, "//", "/")

	return strings.ReplaceAll(path, "//", "/")
}
