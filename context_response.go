package chi

import (
	"bytes"
	"net/http"

	"github.com/go-rat/chix"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
)

type ContextResponse struct {
	ctx    *Context
	render *chix.Render
	origin contractshttp.ResponseOrigin
}

func NewContextResponse(ctx *Context, origin contractshttp.ResponseOrigin) *ContextResponse {
	return &ContextResponse{ctx, chix.NewRender(ctx.w, ctx.r), origin}
}

func (r *ContextResponse) Cookie(cookie contractshttp.Cookie) contractshttp.ContextResponse {
	if cookie.MaxAge == 0 {
		if !cookie.Expires.IsZero() {
			cookie.MaxAge = int(cookie.Expires.Sub(carbon.Now().StdTime()).Seconds())
		}
	}

	sameSiteOptions := map[string]http.SameSite{
		"strict": http.SameSiteStrictMode,
		"lax":    http.SameSiteLaxMode,
		"none":   http.SameSiteNoneMode,
	}

	var sameSite http.SameSite
	if val, ok := sameSiteOptions[cookie.SameSite]; ok {
		sameSite = val
	} else {
		sameSite = http.SameSiteDefaultMode
	}

	r.render.Cookie(&http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		MaxAge:   cookie.MaxAge,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
		SameSite: sameSite,
	})

	return r
}

func (r *ContextResponse) Data(code int, contentType string, data []byte) contractshttp.Response {
	return &DataResponse{code, contentType, data, r.render}
}

func (r *ContextResponse) Download(filepath, filename string) contractshttp.Response {
	return &DownloadResponse{filename, filepath, r.render}
}

func (r *ContextResponse) File(filepath string) contractshttp.Response {
	return &FileResponse{filepath, r.render}
}

func (r *ContextResponse) Header(key, value string) contractshttp.ContextResponse {
	r.render.Header(key, value)

	return r
}

func (r *ContextResponse) Json(code int, obj any) contractshttp.Response {
	return &JsonResponse{code, obj, r.render}
}

func (r *ContextResponse) NoContent(code ...int) contractshttp.Response {
	if len(code) > 0 {
		return &NoContentResponse{code[0], r.render}
	}

	return &NoContentResponse{http.StatusNoContent, r.render}
}

func (r *ContextResponse) Origin() contractshttp.ResponseOrigin {
	return r.origin
}

func (r *ContextResponse) Redirect(code int, location string) contractshttp.Response {
	return &RedirectResponse{code, location, r.render}
}

func (r *ContextResponse) String(code int, format string, values ...any) contractshttp.Response {
	return &StringResponse{code, format, r.render, values}
}

func (r *ContextResponse) Success() contractshttp.ResponseStatus {
	return NewStatus(r.render, http.StatusOK)
}

func (r *ContextResponse) Status(code int) contractshttp.ResponseStatus {
	return NewStatus(r.render, code)
}

func (r *ContextResponse) Stream(code int, step func(w contractshttp.StreamWriter) error) contractshttp.Response {
	return &StreamResponse{code, r.render, step}
}

func (r *ContextResponse) View() contractshttp.ResponseView {
	return NewView(r.render)
}

func (r *ContextResponse) WithoutCookie(name string) contractshttp.ContextResponse {
	r.render.WithoutCookie(name)
	return r
}

func (r *ContextResponse) Writer() http.ResponseWriter {
	return r.ctx.w
}

func (r *ContextResponse) Flush() {
	r.render.Flush()
}

type Status struct {
	render *chix.Render
	status int
}

func NewStatus(render *chix.Render, code int) *Status {
	return &Status{render, code}
}

func (r *Status) Data(contentType string, data []byte) contractshttp.Response {
	return &DataResponse{r.status, contentType, data, r.render}
}

func (r *Status) Json(obj any) contractshttp.Response {
	return &JsonResponse{r.status, obj, r.render}
}

func (r *Status) String(format string, values ...any) contractshttp.Response {
	return &StringResponse{r.status, format, r.render, values}
}

func (r *Status) Stream(step func(w contractshttp.StreamWriter) error) contractshttp.Response {
	return &StreamResponse{r.status, r.render, step}
}

func ResponseMiddleware() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		blw := &BodyWriter{body: bytes.NewBufferString("")}
		switch ctx := ctx.(type) {
		case *Context:
			blw.ResponseWriter = ctx.w
			ctx.w = blw
		}

		ctx.WithValue("responseOrigin", blw)
		ctx.Request().Next()
	}
}

type BodyWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *BodyWriter) Size() int {
	return w.body.Len()
}

func (w *BodyWriter) Status() int {
	return w.status
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func (w *BodyWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *BodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)

	return w.ResponseWriter.Write([]byte(s))
}

func (w *BodyWriter) Body() *bytes.Buffer {
	return w.body
}

func (w *BodyWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
