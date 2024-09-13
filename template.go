package chi

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

type Delims struct {
	Left  string
	Right string
}

type RenderOptions struct {
	Delims  *Delims
	FuncMap template.FuncMap
}

func NewTemplate(options RenderOptions) (*template.Template, error) {
	instance := template.New("")
	if options.Delims != nil {
		instance.Delims(options.Delims.Left, options.Delims.Right)
	}
	if options.FuncMap != nil {
		instance.Funcs(options.FuncMap)
	}

	dir := "resources/views"
	if support.RelativePath != "" {
		dir = support.RelativePath + "/" + dir
	}

	if !file.Exists(dir) {
		return nil, nil
	}

	var files []string
	if err := filepath.Walk(dir, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, fullPath)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	return template.Must(instance.ParseFiles(files...)), nil
}

// DefaultTemplate creates a TemplateRender instance with default options.
func DefaultTemplate() (*template.Template, error) {
	return NewTemplate(RenderOptions{})
}
