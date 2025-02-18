package catalog

import (
	"context"
	"embed"
	"io/fs"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/catalog/render"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate views/...

//go:embed static
var staticFilesFS embed.FS

type Plugin struct {
	core.NoOpShutdown
}

func (Plugin) Setup(_ context.Context, deps core.Dependencies) error {
	staticFiles, err := fs.Sub(staticFilesFS, "static")
	if err != nil {
		return err
	}

	router := deps.Router()

	router.StaticFS("static", staticFiles)

	favicon, err := fs.Sub(staticFiles, "favicon.ico")
	if err != nil {
		return err
	}

	router.StaticFS("favicon.ico", favicon)

	entityHandlers := render.Handlers{Dependencies: deps}
	router.GET("/", entityHandlers.Index)

	catalogGroup := router.Group("/catalog")
	catalogGroup.GET("/view/entities/kind/:kind", entityHandlers.ViewEntities)
	catalogGroup.GET("/view/entity/fullname/:fullname", entityHandlers.ViewEntity)

	catalogGroup.GET("/component/entities/kind/:kind", entityHandlers.ListEntities)
	catalogGroup.GET("/component/entity/fullname/:fullname", entityHandlers.ContentEntity)
	catalogGroup.GET("/component/entity/info/fullname/:fullname", entityHandlers.ContentEntityInfo)

	return nil
}
