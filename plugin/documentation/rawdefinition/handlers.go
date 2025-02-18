package rawdefinition

import (
	"errors"
	"io"
	"strconv"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	core.Dependencies
}

var errInvalidDefinitionType = errors.New("definition is not a string")

func (h Handlers) RawDefinitionHandler(reqCtx echo.Context) error {
	fullname := reqCtx.Param("fullname")

	preformatted, _ := strconv.ParseBool(reqCtx.QueryParam("htmlPreformatted"))
	if preformatted {
		_, writeErr := io.WriteString(reqCtx.Response().Writer, "<pre>")
		if writeErr != nil {
			return writeErr
		}

		defer func() {
			_, _ = io.WriteString(reqCtx.Response().Writer, "</pre>")
		}()
	}

	entity, err := h.Repo().GetByName(reqCtx.Request().Context(), fullname)
	if err != nil {
		return err
	}

	if txt, ok := entity.Spec.Definition.(string); !ok {
		return errInvalidDefinitionType
	} else {
		_, err = io.WriteString(reqCtx.Response().Writer, txt)
	}

	return err
}
