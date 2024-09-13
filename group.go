package chi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
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
	r.getRoutesWithMiddlewares().Handle(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Get(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Get(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Post(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Post(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Delete(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Delete(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Patch(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Patch(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Put(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Put(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Options(relativePath string, handler httpcontract.HandlerFunc) {
	r.getRoutesWithMiddlewares().Options(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) Resource(relativePath string, controller httpcontract.ResourceController) {
	r.getRoutesWithMiddlewares().Get(relativePath, handlerToChiHandler(r.instance, controller.Index))
	r.getRoutesWithMiddlewares().Post(relativePath, handlerToChiHandler(r.instance, controller.Store))
	r.getRoutesWithMiddlewares().Get(relativePath+"/{id}", handlerToChiHandler(r.instance, controller.Show))
	r.getRoutesWithMiddlewares().Put(relativePath+"/{id}", handlerToChiHandler(r.instance, controller.Update))
	r.getRoutesWithMiddlewares().Patch(relativePath+"/{id}", handlerToChiHandler(r.instance, controller.Update))
	r.getRoutesWithMiddlewares().Delete(relativePath+"/{id}", handlerToChiHandler(r.instance, controller.Destroy))
	r.clearMiddlewares()
}

func (r *Group) Static(relativePath, root string) {
	r.StaticFS(relativePath, http.Dir(root))
	r.clearMiddlewares()
}

func (r *Group) StaticFile(relativePath, filepath string) {
	handler := httpcontract.HandlerFunc(func(ctx httpcontract.Context) httpcontract.Response {
		return ctx.Response().File(filepath)
	})

	r.getRoutesWithMiddlewares().Get(relativePath, handlerToChiHandler(r.instance, handler))
	r.getRoutesWithMiddlewares().Head(relativePath, handlerToChiHandler(r.instance, handler))
	r.clearMiddlewares()
}

func (r *Group) StaticFS(relativePath string, fs http.FileSystem) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	r.getRoutesWithMiddlewares().Handle(relativePath, fileServer)
	r.clearMiddlewares()
}

func (r *Group) getRoutesWithMiddlewares() chi.Router {
	prefix := r.originPrefix + "/" + r.prefix

	r.prefix = ""
	group := r.instance.mux.Route(prefix, func(r chi.Router) {}) // We don't need to add any handler here

	var middlewares []func(http.Handler) http.Handler
	ginOriginMiddlewares := middlewaresToChiHandlers(r.instance, r.originMiddlewares)
	ginMiddlewares := middlewaresToChiHandlers(r.instance, r.middlewares)
	ginLastMiddlewares := middlewaresToChiHandlers(r.instance, r.lastMiddlewares)

	middlewares = append(middlewares, ginOriginMiddlewares...)
	middlewares = append(middlewares, ginMiddlewares...)
	middlewares = append(middlewares, ginLastMiddlewares...)

	if len(middlewares) > 0 {
		group.Use(middlewares...)
	}

	return group
}

func (r *Group) clearMiddlewares() {
	r.middlewares = []httpcontract.Middleware{}
}
