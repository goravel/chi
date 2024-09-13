package chi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-rat/chix"
	"github.com/go-rat/chix/binder"
	"github.com/gookit/validate"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/log"
	contractsession "github.com/goravel/framework/contracts/session"
	contractsvalidate "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/support/json"
	"github.com/goravel/framework/validation"
	"github.com/spf13/cast"
)

type ContextRequest struct {
	ctx        *Context
	bind       *chix.Bind
	render     *chix.Render
	httpBody   map[string]any
	log        log.Log
	validation contractsvalidate.Validation
}

func NewContextRequest(ctx *Context, log log.Log, validation contractsvalidate.Validation) contractshttp.ContextRequest {
	httpBody, err := getHttpBody(ctx)
	if err != nil {
		LogFacade.Error(fmt.Sprintf("%+v", errors.Unwrap(err)))
	}

	return &ContextRequest{ctx: ctx, bind: chix.NewBind(ctx.r), render: chix.NewRender(ctx.w, ctx.r), httpBody: httpBody, log: log, validation: validation}
}

func (r *ContextRequest) AbortWithStatus(code int) {
	r.render.Status(code)
}

func (r *ContextRequest) AbortWithStatusJson(code int, jsonObj any) {
	r.render.Status(code)
	r.render.JSON(jsonObj)
}

func (r *ContextRequest) All() map[string]any {
	var (
		dataMap  = make(map[string]any)
		queryMap = make(map[string]any)
	)

	for key, query := range r.ctx.r.URL.Query() {
		queryMap[key] = strings.Join(query, ",")
	}

	chiCtx := chi.RouteContext(r.ctx.r.Context())
	for k := len(chiCtx.URLParams.Keys) - 1; k >= 0; k-- {
		key := chiCtx.URLParams.Keys[k]
		val := chiCtx.URLParams.Values[k]
		if key == "*" {
			continue
		}
		dataMap[key] = val
	}
	for k, v := range queryMap {
		dataMap[k] = v
	}
	for k, v := range r.httpBody {
		dataMap[k] = v
	}

	return dataMap
}

func (r *ContextRequest) Bind(obj any) error {
	return r.bind.Body(obj)
}

func (r *ContextRequest) BindQuery(obj any) error {
	return r.bind.Query(obj)
}

func (r *ContextRequest) Cookie(key string, defaultValue ...string) string {
	for _, cookie := range r.ctx.r.Cookies() {
		if cookie.Name == key {
			return cookie.Value
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

func (r *ContextRequest) Form(key string, defaultValue ...string) string {
	// TODO optimize performance
	form := make(map[string]string)
	if err := r.bind.Form(&form); err == nil {
		if value, exist := form[key]; exist {
			return value
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

func (r *ContextRequest) File(name string) (contractsfilesystem.File, error) {
	if err := r.ctx.r.ParseMultipartForm(r.ctx.instance.maxMultipartMemory); err != nil {
		return nil, err
	}
	f, fh, err := r.ctx.r.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return filesystem.NewFileFromRequest(fh)
}

func (r *ContextRequest) FullUrl() string {
	prefix := "https://"
	if r.ctx.r.TLS == nil {
		prefix = "http://"
	}

	if r.ctx.r.Host == "" {
		return ""
	}

	return prefix + r.ctx.r.Host + r.ctx.r.RequestURI
}

func (r *ContextRequest) Header(key string, defaultValue ...string) string {
	header := r.ctx.r.Header.Get(key)
	if header != "" {
		return header
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *ContextRequest) Headers() http.Header {
	return r.ctx.r.Header
}

func (r *ContextRequest) Host() string {
	return r.ctx.r.Host
}

func (r *ContextRequest) HasSession() bool {
	_, ok := r.ctx.Value("session").(contractsession.Session)
	return ok
}

func (r *ContextRequest) Json(key string, defaultValue ...string) string {
	data := make(map[string]any)
	if err := r.Bind(&data); err != nil {
		if len(defaultValue) == 0 {
			return ""
		} else {
			return defaultValue[0]
		}
	}

	if value, exist := data[key]; exist {
		return cast.ToString(value)
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}

func (r *ContextRequest) Method() string {
	return r.ctx.r.Method
}

func (r *ContextRequest) Next() {
	// TODO how to implement this?
}

func (r *ContextRequest) Query(key string, defaultValue ...string) string {
	// TODO optimize performance
	query := make(map[string]string)
	if err := r.bind.Query(&query); err == nil {
		if value, exist := query[key]; exist {
			return value
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

func (r *ContextRequest) QueryInt(key string, defaultValue ...int) int {
	if val := r.Query(key); val != "" {
		return cast.ToInt(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *ContextRequest) QueryInt64(key string, defaultValue ...int64) int64 {
	if val := r.Query(key); val != "" {
		return cast.ToInt64(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

func (r *ContextRequest) QueryBool(key string, defaultValue ...bool) bool {
	if val := r.Query(key); val != "" {
		return stringToBool(val)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

func (r *ContextRequest) QueryArray(key string) []string {
	// TODO optimize performance
	queries := make(map[string][]string)
	if err := r.bind.Query(&queries); err == nil {
		if value, exist := queries[key]; exist {
			return value
		}
	}

	return nil
}

func (r *ContextRequest) QueryMap(key string) map[string]string {
	// TODO optimize performance
	queries := make(map[string][]string)
	if err := r.bind.Query(&queries); err != nil {
		return nil
	}

	// The bind.Query() method will bind query foo[bar]=baz to map["foo.bar"] = baz
	key = key + "."
	data := make(map[string]string)
	for k, v := range queries {
		if strings.HasPrefix(k, key) {
			k = strings.TrimPrefix(k, key)
			data[k] = strings.Join(v, ",")
		}
	}

	return data
}

func (r *ContextRequest) Queries() map[string]string {
	queries := make(map[string]string)

	for key, query := range r.ctx.r.URL.Query() {
		queries[key] = strings.Join(query, ",")
	}

	return queries
}

func (r *ContextRequest) Origin() *http.Request {
	return r.ctx.r
}

func (r *ContextRequest) Path() string {
	return r.ctx.r.URL.Path
}

func (r *ContextRequest) Input(key string, defaultValue ...string) string {
	valueFromHttpBody := r.getValueFromHttpBody(key)
	if valueFromHttpBody != nil {
		switch reflect.ValueOf(valueFromHttpBody).Kind() {
		case reflect.Map:
			valueFromHttpBodyObByte, err := json.Marshal(valueFromHttpBody)
			if err != nil {
				return ""
			}

			return string(valueFromHttpBodyObByte)
		case reflect.Slice:
			return strings.Join(cast.ToStringSlice(valueFromHttpBody), ",")
		default:
			return cast.ToString(valueFromHttpBody)
		}
	}

	if r.Query(key) != "" {
		return r.Query(key)
	}

	chiCtx := chi.RouteContext(r.ctx.r.Context())
	value := chiCtx.URLParam(key)
	if len(value) == 0 && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return value
}

func (r *ContextRequest) InputArray(key string, defaultValue ...[]string) []string {
	if valueFromHttpBody := r.getValueFromHttpBody(key); valueFromHttpBody != nil {
		return cast.ToStringSlice(valueFromHttpBody)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return []string{}
	}
}

func (r *ContextRequest) InputMap(key string, defaultValue ...map[string]string) map[string]string {
	if valueFromHttpBody := r.getValueFromHttpBody(key); valueFromHttpBody != nil {
		return cast.ToStringMapString(valueFromHttpBody)
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	} else {
		return map[string]string{}
	}
}

func (r *ContextRequest) InputInt(key string, defaultValue ...int) int {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt(value)
}

func (r *ContextRequest) InputInt64(key string, defaultValue ...int64) int64 {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return cast.ToInt64(value)
}

func (r *ContextRequest) InputBool(key string, defaultValue ...bool) bool {
	value := r.Input(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return stringToBool(value)
}

func (r *ContextRequest) Ip() string {
	ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.ctx.r.RemoteAddr))
	return ip
}

func (r *ContextRequest) Route(key string) string {
	chiCtx := chi.RouteContext(r.ctx.r.Context())
	return chiCtx.URLParam(key)
}

func (r *ContextRequest) RouteInt(key string) int {
	chiCtx := chi.RouteContext(r.ctx.r.Context())
	val := chiCtx.URLParam(key)

	return cast.ToInt(val)
}

func (r *ContextRequest) RouteInt64(key string) int64 {
	chiCtx := chi.RouteContext(r.ctx.r.Context())
	val := chiCtx.URLParam(key)

	return cast.ToInt64(val)
}

func (r *ContextRequest) Session() contractsession.Session {
	s, ok := r.ctx.Value("session").(contractsession.Session)
	if !ok {
		return nil
	}
	return s
}

func (r *ContextRequest) SetSession(session contractsession.Session) contractshttp.ContextRequest {
	r.ctx.WithValue("session", session)

	return r
}

func (r *ContextRequest) Url() string {
	return r.ctx.r.RequestURI
}

func (r *ContextRequest) Validate(rules map[string]string, options ...contractsvalidate.Option) (contractsvalidate.Validator, error) {
	if len(rules) == 0 {
		return nil, errors.New("rules can't be empty")
	}

	options = append(options, validation.Rules(rules), validation.CustomRules(r.validation.Rules()), validation.CustomFilters(r.validation.Filters()))

	dataFace, err := validate.FromRequest(r.ctx.Request().Origin())
	if err != nil {
		return nil, err
	}

	for key, query := range r.ctx.r.URL.Query() {
		if _, exist := dataFace.Get(key); !exist {
			if _, err := dataFace.Set(key, strings.Join(query, ",")); err != nil {
				return nil, err
			}
		}
	}

	chiCtx := chi.RouteContext(r.ctx.r.Context())
	for k := len(chiCtx.URLParams.Keys) - 1; k >= 0; k-- {
		key := chiCtx.URLParams.Keys[k]
		val := chiCtx.URLParams.Values[k]
		if key == "*" {
			continue
		}
		if _, exist := dataFace.Get(key); !exist {
			if _, err := dataFace.Set(key, val); err != nil {
				return nil, err
			}
		}
	}

	return r.validation.Make(dataFace, rules, options...)
}

func (r *ContextRequest) ValidateRequest(request contractshttp.FormRequest) (contractsvalidate.Errors, error) {
	if err := request.Authorize(r.ctx); err != nil {
		return nil, err
	}

	validator, err := r.Validate(request.Rules(r.ctx), validation.Filters(request.Filters(r.ctx)), validation.Messages(request.Messages(r.ctx)), validation.Attributes(request.Attributes(r.ctx)), func(options map[string]any) {
		options["prepareForValidation"] = request.PrepareForValidation
	})
	if err != nil {
		return nil, err
	}

	if err := validator.Bind(request); err != nil {
		return nil, err
	}

	return validator.Errors(), nil
}

func (r *ContextRequest) getValueFromHttpBody(key string) any {
	if r.httpBody == nil {
		return nil
	}

	var current any
	current = r.httpBody
	keys := strings.Split(key, ".")
	for _, k := range keys {
		currentValue := reflect.ValueOf(current)
		switch currentValue.Kind() {
		case reflect.Map:
			if value := currentValue.MapIndex(reflect.ValueOf(k)); value.IsValid() {
				current = value.Interface()
			} else {
				if value := currentValue.MapIndex(reflect.ValueOf(k + "[]")); value.IsValid() {
					current = value.Interface()
				} else {
					return nil
				}
			}
		case reflect.Slice:
			if number, err := strconv.Atoi(k); err == nil {
				return cast.ToStringSlice(current)[number]
			} else {
				return nil
			}
		}
	}

	return current
}

func getHttpBody(ctx *Context) (map[string]any, error) {
	if ctx.r == nil || ctx.r.Body == nil || ctx.r.ContentLength == 0 {
		return nil, nil
	}

	contentType := strings.ToLower(ctx.r.Header.Get("Content-Type"))
	contentType = binder.FilterFlags(contentType)
	data := make(map[string]any)
	if contentType == "application/json" {
		bodyBytes, err := io.ReadAll(ctx.r.Body)
		_ = ctx.r.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("retrieve json error: %v", err)
		}

		if err = json.Unmarshal(bodyBytes, &data); err != nil {
			return nil, fmt.Errorf("decode json [%v] error: %v", string(bodyBytes), err)
		}

		ctx.r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if contentType == "multipart/form-data" {
		if ctx.r.PostForm == nil {
			const defaultMemory = 32 << 20
			if err := ctx.r.ParseMultipartForm(defaultMemory); err != nil {
				return nil, fmt.Errorf("parse multipart form error: %v", err)
			}
		}
		for k, v := range ctx.r.PostForm {
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
		for k, v := range ctx.r.MultipartForm.File {
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
	}

	if contentType == "application/x-www-form-urlencoded" {
		if ctx.r.PostForm == nil {
			if err := ctx.r.ParseForm(); err != nil {
				return nil, fmt.Errorf("parse form error: %v", err)
			}
		}
		for k, v := range ctx.r.PostForm {
			if len(v) > 1 {
				data[k] = v
			} else if len(v) == 1 {
				data[k] = v[0]
			}
		}
	}

	return data, nil
}

func stringToBool(value string) bool {
	return value == "1" || value == "true" || value == "on" || value == "yes"
}
