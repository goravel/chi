package chi

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"

	contractshttp "github.com/goravel/framework/contracts/http"
)

type View struct {
	template *template.Template
	w        http.ResponseWriter
}

func NewView(template *template.Template, w http.ResponseWriter) *View {
	return &View{template, w}
}

func (receive *View) Make(view string, data ...any) contractshttp.Response {
	shared := ViewFacade.GetShared()
	if len(data) == 0 {
		return &HtmlResponse{shared, view, receive.template, receive.w}
	} else {
		dataType := reflect.TypeOf(data[0])
		switch dataType.Kind() {
		case reflect.Struct:
			dataMap := structToMap(data[0])
			for key, value := range dataMap {
				shared[key] = value
			}

			return &HtmlResponse{shared, view, receive.template, receive.w}
		case reflect.Map:
			fillShared(data[0], shared)

			return &HtmlResponse{shared, view, receive.template, receive.w}
		default:
			panic(fmt.Sprintf("make %s view failed, data must be map or struct", view))
		}
	}
}

func (receive *View) First(views []string, data ...any) contractshttp.Response {
	for _, view := range views {
		if ViewFacade.Exists(view) {
			return receive.Make(view, data...)
		}
	}

	panic("no view exists")
}

func structToMap(data any) map[string]any {
	res := make(map[string]any)
	modelType := reflect.TypeOf(data)
	modelValue := reflect.ValueOf(data)

	if modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
		modelValue = modelValue.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		if !modelType.Field(i).IsExported() {
			continue
		}
		dbColumn := modelType.Field(i).Name
		if modelValue.Field(i).Kind() == reflect.Pointer {
			if modelValue.Field(i).IsNil() {
				res[dbColumn] = nil
			} else {
				res[dbColumn] = modelValue.Field(i).Elem().Interface()
			}
		} else {
			res[dbColumn] = modelValue.Field(i).Interface()
		}
	}

	return res
}

func fillShared(data any, shared map[string]any) {
	dataValue := reflect.ValueOf(data)
	keys := dataValue.MapKeys()
	for key, value := range shared {
		exist := false
		for _, k := range keys {
			if k.String() == key {
				exist = true
				break
			}
		}
		if !exist {
			dataValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		}
	}
}
