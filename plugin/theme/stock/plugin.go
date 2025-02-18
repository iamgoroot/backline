package stock

import (
	"context"
	"embed"
	"io"
	"io/fs"

	"github.com/iamgoroot/backline/pkg/core"
)

//go:embed css
var staticFilesFS embed.FS

type Theme struct {
	core.NoOpShutdown
}

func (Theme) Setup(_ context.Context, deps core.Dependencies) error {
	staticFiles, err := fs.Sub(staticFilesFS, "css")
	if err != nil {
		return err
	}

	deps.Router().StaticFS("static/css", staticFiles)

	return nil
}

func (Theme) HTMLHeader() core.Component {
	return htmlHeaders
}

var htmlHeaders = core.ComponentFunc(func(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, `
	<script src="/static/lib/htmx.js"></script>
	<link rel="stylesheet" href="/static/css/sidebar.css"/>
	<link rel="stylesheet" href="/static/css/catalog.css"/>
	<link rel="stylesheet" href="/static/css/entity.css"/>
	<link rel="stylesheet" href="/static/css/theme.css"/>
	<link rel="stylesheet" href="/static/css/search.css"/>
	`)
	return err
}, 0)
