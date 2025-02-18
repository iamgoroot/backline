package asyncapi

import (
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/documentation/asyncapi/views"
	"github.com/labstack/echo/v4"
)

type handlers struct {
	core.Dependencies
}

func (h handlers) entityTabHandler(reqCtx echo.Context) error {
	fullname := reqCtx.Param("fullname")

	entity, err := h.Repo().GetByName(reqCtx.Request().Context(), fullname)
	if err != nil {
		return err
	}

	return views.EntityTab(entity).Render(reqCtx.Request().Context(), reqCtx.Response().Writer)
}
