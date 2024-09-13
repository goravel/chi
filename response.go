package chi

import (
	"html/template"
	"io"
	"net/http"

	"github.com/go-rat/chix"
	contractshttp "github.com/goravel/framework/contracts/http"
)

type DataResponse struct {
	code        int
	contentType string
	data        []byte
	render      *chix.Render
}

func (r *DataResponse) Render() error {
	r.render.Status(r.code)
	r.render.ContentType(r.contentType)
	r.render.Data(r.data)

	return nil
}

type DownloadResponse struct {
	filename string
	filepath string
	render   *chix.Render
}

func (r *DownloadResponse) Render() error {
	r.render.Download(r.filepath, r.filename)

	return nil
}

type FileResponse struct {
	filepath string
	render   *chix.Render
}

func (r *FileResponse) Render() error {
	r.render.File(r.filepath)

	return nil
}

type JsonResponse struct {
	code   int
	obj    any
	render *chix.Render
}

func (r *JsonResponse) Render() error {
	r.render.Status(r.code)
	r.render.JSON(r.obj)

	return nil
}

type NoContentResponse struct {
	code   int
	render *chix.Render
}

func (r *NoContentResponse) Render() error {
	r.render.Status(r.code)

	return nil
}

type RedirectResponse struct {
	code     int
	location string
	render   *chix.Render
}

func (r *RedirectResponse) Render() error {
	if r.code == http.StatusMovedPermanently {
		r.render.RedirectPermanent(r.location)
	} else {
		r.render.Redirect(r.location)
	}

	return nil
}

type StringResponse struct {
	code   int
	format string
	render *chix.Render
	values []any
}

func (r *StringResponse) Render() error {
	r.render.Status(r.code)

	if len(r.values) == 0 {
		r.render.PlainText(r.format)
		return nil
	}

	r.render.Header(chix.HeaderContentType, r.format)
	r.render.PlainText(r.values[0].(string))

	return nil
}

type HtmlResponse struct {
	data     any
	view     string
	template *template.Template
	w        http.ResponseWriter
}

func (r *HtmlResponse) Render() error {
	r.w.Header().Set("Content-Type", chix.MIMETextHTMLCharsetUTF8)
	if r.view == "" {
		return r.template.Execute(r.w, r.data)
	}

	return r.template.ExecuteTemplate(r.w, r.view, r.data)
}

type StreamResponse struct {
	code   int
	render *chix.Render
	writer func(w contractshttp.StreamWriter) error
}

func (r *StreamResponse) Render() error {
	r.render.Status(r.code)
	r.render.Stream(func(w io.Writer) bool {
		if err := r.writer(NewStreamWriter(r.render, w)); err != nil {
			return false
		}
		return true
	})

	return nil
}
