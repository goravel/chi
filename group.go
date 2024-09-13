package chi

import (
	"net/http"
	"strings"

	"github.com/goravel/framework/contracts/config"
	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
)

type Group struct {
	config            config.Config
	instance          *Instance
	originPrefix      string
	prefix            string
	originMiddlewares []httpcontract.Middleware
	middlewares       []httpcontract.Middleware
	lastMiddlewares   []httpcontract.Middleware
}

func NewGroup(config config.Config, instance *Instance, prefix string, originMiddlewares []httpcontract.Middleware, lastMiddlewares []httpcontract.Middleware) route.Router {
	return &Group{
		config:            config,
		instance:          instance,
		originPrefix:      prefix,
		originMiddlewares: originMiddlewares,
		lastMiddlewares:   lastMiddlewares,
	}
}

func (r *Group) Group(handler route.GroupFunc) {
	var middlewares []httpcontract.Middleware
	middlewares = append(middlewares, r.originMiddlewares...)
	middlewares = append(middlewares, r.middlewares...)
	r.middlewares = []httpcontract.Middleware{}
	prefix := r.originPrefix + "/" + r.prefix
	r.prefix = ""

	handler(NewGroup(r.config, r.instance, prefix, middlewares, r.lastMiddlewares))
}

func (r *Group) Prefix(addr string) route.Router {
	r.prefix += "/" + addr

	return r
}

func (r *Group) Middleware(middlewares ...httpcontract.Middleware) route.Router {
	r.middlewares = append(r.middlewares, middlewares...)

	return r
}

func (r *Group) Any(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Handle(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Get(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Post(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Delete(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Patch(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Put(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.instance.mux.With(r.getMiddlewares()...).Options(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Resource(relativePath string, controller httpcontract.ResourceController) {
	r.instance.mux.With(r.getMiddlewares()...).Get(r.getPath(relativePath), handlerToChiHandler(r.instance, controller.Index))
	r.instance.mux.With(r.getMiddlewares()...).Post(r.getPath(relativePath), handlerToChiHandler(r.instance, controller.Store))
	r.instance.mux.With(r.getMiddlewares()...).Get(r.getPath(relativePath)+"/{id}", handlerToChiHandler(r.instance, controller.Show))
	r.instance.mux.With(r.getMiddlewares()...).Put(r.getPath(relativePath)+"/{id}", handlerToChiHandler(r.instance, controller.Update))
	r.instance.mux.With(r.getMiddlewares()...).Patch(r.getPath(relativePath)+"/{id}", handlerToChiHandler(r.instance, controller.Update))
	r.instance.mux.With(r.getMiddlewares()...).Delete(r.getPath(relativePath)+"/{id}", handlerToChiHandler(r.instance, controller.Destroy))
	r.clearMiddlewares()
}

func (r *Group) Static(relativePath, root string) {
	r.StaticFS(r.getPath(relativePath), http.Dir(root))
	r.clearMiddlewares()
}

func (r *Group) StaticFile(relativePath, filepath string) {
	handler := httpcontract.HandlerFunc(func(ctx httpcontract.Context) httpcontract.Response {
		return ctx.Response().File(filepath)
	})

	r.instance.mux.With(r.getMiddlewares()...).Get(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.instance.mux.With(r.getMiddlewares()...).Head(r.getPath(relativePath), handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) StaticFS(relativePath string, fs http.FileSystem) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	fileServer := http.StripPrefix(r.getPath(relativePath), http.FileServer(fs))
	r.instance.mux.With(r.getMiddlewares()...).Handle(r.getPath(relativePath), fileServer)
	r.clearMiddlewares()
}

func (r *Group) getPath(relativePath string) string {
	path := r.originPrefix + "/" + r.prefix + "/" + relativePath
	path = mergeSlashForPath(path)
	r.prefix = ""
	return path
}

func (r *Group) getMiddlewares() []func(http.Handler) http.Handler {
	var middlewares []func(http.Handler) http.Handler
	middlewares = append(middlewares, middlewaresToChiHandlers(r.instance, r.originMiddlewares)...)
	middlewares = append(middlewares, middlewaresToChiHandlers(r.instance, r.middlewares)...)
	middlewares = append(middlewares, middlewaresToChiHandlers(r.instance, r.lastMiddlewares)...)

	return middlewares
}

func (r *Group) clearMiddlewares() {
	r.middlewares = []httpcontract.Middleware{}
}
