package chi

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/http"
)

func Background() http.Context {
	return &Context{
		r: &nethttp.Request{},
	}
}

type Context struct {
	r        *nethttp.Request
	w        nethttp.ResponseWriter
	instance *Instance
	request  http.ContextRequest
	response http.ContextResponse
}

func NewContext(instance *Instance, w nethttp.ResponseWriter, r *nethttp.Request) http.Context {
	return &Context{r: r, w: w, instance: instance}
}

func (c *Context) Request() http.ContextRequest {
	if c.request == nil {
		c.request = NewContextRequest(c, LogFacade, ValidationFacade)
	}

	return c.request
}

func (c *Context) Response() http.ContextResponse {
	if c.response == nil {
		c.response = NewContextResponse(c, &BodyWriter{ResponseWriter: c.w})
	}

	return c.response
}

func (c *Context) WithValue(key string, value any) {
	c.r = c.r.WithContext(context.WithValue(c.r.Context(), key, value)) //nolint:staticcheck
}

func (c *Context) Context() context.Context {
	return c.r.Context()
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.r.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.r.Context().Done()
}

func (c *Context) Err() error {
	return c.r.Context().Err()
}

func (c *Context) Value(key any) any {
	return c.r.Context().Value(key)
}

func (c *Context) Instance() *Instance {
	return c.instance
}
