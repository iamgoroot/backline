package render

import (
	"github.com/a-h/templ"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/catalog/views"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	core.Dependencies
}

func (h Handlers) Index(c echo.Context) error {
	return Index(h.Plugins(), c, nil)
}
func (h Handlers) ViewEntities(c echo.Context) error {
	table := safeTempl(h.renderEntitiesTable(c))
	return Index(h.Plugins(), c, table)
}

func Index(plugins core.Plugins, c echo.Context, content templ.Component) error {
	index := views.Index(plugins, content)
	return renderEcho(c, index)
}
