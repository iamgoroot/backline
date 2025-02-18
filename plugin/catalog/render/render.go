package render

import (
	"github.com/a-h/templ"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
)

func renderEcho(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func safeTempl(component templ.Component, err error) templ.Component {
	if err != nil {
		return core.ErrComponent(err)
	}

	return component
}
