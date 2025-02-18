package render

import (
	"github.com/a-h/templ"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/plugin/catalog/views/catalog"
	"github.com/labstack/echo/v4"
)

func (h Handlers) ListEntities(c echo.Context) error {
	table := safeTempl(h.renderEntitiesTable(c))
	return renderEcho(c, table)
}

func (h Handlers) renderEntitiesTable(reqCtx echo.Context) (templ.Component, error) {
	req := &model.ListEntityReq{}
	if err := reqCtx.Bind(req); err != nil {
		return nil, err
	}

	entities, err := h.Repo().List(reqCtx.Request().Context(), req)
	if err != nil {
		return nil, err
	}

	return catalog.EntityRows(entities), nil
}
