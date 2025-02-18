package rawdefinition

import (
	"context"
	"github.com/iamgoroot/backline/plugin/documentation/rawdefinition/internal/views"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate internal/views/...

var _ core.EntityTabPlugin = Plugin{}

type Plugin struct {
	core.NoOpShutdown
}

func (p Plugin) Setup(_ context.Context, deps core.Dependencies) error {
	h := Handlers{Dependencies: deps}
	deps.Router().Any("rawdefinition/by-fullname/:fullname", h.RawDefinitionHandler)

	return nil
}

func (Plugin) EntityTabLink(entity *model.Entity) core.Component {
	if entity.Spec.Definition == nil || !strings.EqualFold(entity.Kind, "api") {
		return core.EmptyComponent{}
	}
	return core.WeighedComponent{Component: views.TabLink(entity.FullName)}
}
