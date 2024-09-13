# Chi

[![Doc](https://pkg.go.dev/badge/github.com/goravel/chi)](https://pkg.go.dev/github.com/goravel/chi)
[![Go](https://img.shields.io/github/go-mod/go-version/goravel/chi)](https://go.dev/)
[![Release](https://img.shields.io/github/release/goravel/chi.svg)](https://github.com/goravel/chi/releases)
[![Test](https://github.com/goravel/chi/actions/workflows/test.yml/badge.svg)](https://github.com/goravel/chi/actions)
[![Report Card](https://goreportcard.com/badge/github.com/goravel/chi)](https://goreportcard.com/report/github.com/goravel/chi)
[![Codecov](https://codecov.io/gh/goravel/chi/branch/master/graph/badge.svg)](https://codecov.io/gh/goravel/chi)
![License](https://img.shields.io/github/license/goravel/chi)

Chi http driver for Goravel.

## Version

| goravel/chi | goravel/framework |
|-------------|-------------------|
| v1.0.x      | v1.15.x           |

## Install

1. Add package

```
go get -u github.com/goravel/chi
```

2. Register service provider

```
// config/app.go
import "github.com/goravel/chi"

"providers": []foundation.ServiceProvider{
    ...
    &chi.ServiceProvider{},
}
```

3. Add chi config to `config/http.go` file

```
// config/http.go
import (
    "html/template"

    chifacades "github.com/goravel/chi/facades"
    "github.com/goravel/chi"
)

"default": "chi",

"drivers": map[string]any{
    "chi": map[string]any{
        // Optional, default is 4096 KB
        "body_limit": 4096,
        "header_limit": 4096,
        "route": func() (route.Route, error) {
            return chifacades.Route(), nil
        },
        // Optional, default is http/template
        "template": func() (*template.Template, error) {
            return new(template.Template), nil
        },
    },
},
```

## Testing

Run command below to run test:

```
go test ./...
```
