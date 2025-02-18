package asyncapi

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/plugin/documentation/asyncapi/views"
	"github.com/iamgoroot/backline/plugin/documentation/rawdefinition"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate views/...

var _ core.EntityTabPlugin = Plugin{}

type Plugin struct {
	core.NoOpShutdown
}

func (p Plugin) Setup(_ context.Context, deps core.Dependencies) error {
	deps.Router().GET("asyncapi-viewer/render/by-fullname/:fullname", handlers{Dependencies: deps}.entityTabHandler)
	// mount raw definition endpoint for reading spec
	deps.Router().GET(
		"asyncapi-viewer/spec/by-fullname/:fullname",
		rawdefinition.Handlers{Dependencies: deps}.RawDefinitionHandler,
	)

	return nil
}

func (Plugin) EntityTabLink(entity *model.Entity) core.Component {
	return core.ComponentFunc(views.EntityTabLink(entity).Render, 0)
}

func (Plugin) HTMLHeader() core.Component {
	return core.ComponentFunc(views.HTMLHeaders().Render, 0)
}
