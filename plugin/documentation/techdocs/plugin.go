package techdocs

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	"github.com/a-h/templ"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/plugin/catalog/render"
	docsModel "github.com/iamgoroot/backline/plugin/documentation/techdocs/model"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs/views"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate views/...

var _ core.EntityProcessor = &Plugin{}

type Plugin struct {
	core.NoOpShutdown
	core.Dependencies
	Parallelism int
}

func (p *Plugin) Setup(_ context.Context, deps core.Dependencies) error {
	p.Dependencies = deps
	if p.Parallelism == 0 {
		p.Parallelism = runtime.NumCPU()
	}

	deps.Router().GET("/techdocs/component/entity/by-fullname/:fullname", func(c echo.Context) error {
		ctx := c.Request().Context()
		fullName := c.Param("fullname")
		content := p.resolveTechDocsContent(c, fullName)

		return content.Render(ctx, c.Response().Writer)
	})
	deps.Router().GET("/techdocs/view/entity/by-fullname/:fullname", func(reqCtx echo.Context) error {
		ctx := reqCtx.Request().Context()
		fullName := reqCtx.Param("fullname")

		entity, err := p.Repo().GetByName(ctx, fullName)
		if err != nil {
			return err
		}

		content := p.resolveTechDocsContent(reqCtx, fullName)

		return render.ViewEntityContent(p.Plugins(), reqCtx, entity, content)
	})
	deps.Router().GET("/techdocs/item/by-fullname/:fullname/by-item-path/:itemPath", func(reqCtx echo.Context) error {
		ctx := reqCtx.Request().Context()
		fullName := paramDecoded(reqCtx, "fullname")
		itemPath := paramDecoded(reqCtx, "itemPath")

		if itemPath == "" || fullName == "" {
			return reqCtx.String(http.StatusBadRequest, "invalid parameters")
		}

		group := fmt.Sprintf("techdocs-items:%s", fullName)

		var item string

		err := p.StoreKV().Get(ctx, group, itemPath, &item)
		if err != nil {
			return err
		}

		return reqCtx.String(http.StatusOK, item)
	})

	return nil
}

func (p *Plugin) resolveTechDocsContent(c echo.Context, fullName string) templ.Component {
	ctx := c.Request().Context()

	var nav docsModel.Nav

	err := p.StoreKV().Get(ctx, "techdocs-root", fullName, &nav)
	if err != nil {
		return core.ErrComponent(err)
	}

	return views.TechDocs(fullName, &nav)
}

func paramDecoded(c echo.Context, name string) string {
	val, _ := url.QueryUnescape(c.Param(name))
	return val
}

func (p *Plugin) EntityTabLink(entity *model.Entity) core.Component {
	if ref, ok := entity.Metadata.Annotations["backstage.io/techdocs-entity"]; ok {
		return core.WeighedComponent{
			Component: views.DocsLinkByRef(ref),
		}
	}

	if _, ok := entity.Metadata.Annotations["backstage.io/techdocs-ref"]; ok {
		return core.WeighedComponent{
			Component: views.DocsLink(entity.FullName),
		}
	}

	return core.EmptyComponent{}
}
