package render

import (
	"github.com/a-h/templ"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/plugin/catalog/views/catalog"
	"github.com/labstack/echo/v4"
)

func (h Handlers) ViewEntity(reqCtx echo.Context) error {
	entity, err := h.resolveEntity(reqCtx)
	if err != nil {
		return err
	}

	content := catalog.EntityInfo(h.Plugins(), entity)

	return ViewEntityContent(h.Plugins(), reqCtx, entity, content)
}

func (h Handlers) ContentEntity(reqCtx echo.Context) error {
	entity, err := h.resolveEntity(reqCtx)
	if err != nil {
		return err
	}

	info := catalog.EntityInfo(h.Plugins(), entity)
	content := catalog.Entity(h.Plugins(), entity, info)

	return renderEcho(reqCtx, content)
}

func (h Handlers) ContentEntityInfo(reqCtx echo.Context) error {
	entity, err := h.resolveEntity(reqCtx)
	if err != nil {
		return err
	}

	info := catalog.EntityInfo(h.Plugins(), entity)

	return renderEcho(reqCtx, info)
}

func (h Handlers) resolveEntity(c echo.Context) (*model.Entity, error) {
	ctx := c.Request().Context()
	fullname := c.Param("fullname")

	return h.Repo().GetByName(ctx, fullname)
}

func ViewEntityContent(plugins core.Plugins, c echo.Context, entity *model.Entity, content templ.Component) error {
	fullContent := catalog.Entity(plugins, entity, content)
	return Index(plugins, c, fullContent)
}
