package ui

import (
	"context"
	"strconv"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/search/ui/views"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate views/...

type SearchView struct {
	Search core.Search
}

const defaultSearchPageSize = 20

func (plugin *SearchView) Setup(_ context.Context, deps core.Dependencies) error {
	plugin.Search = deps.Search()
	group := deps.Router().Group("/search")

	group.GET("", plugin.searchResultsHandler)

	return nil
}

func (plugin *SearchView) searchResultsHandler(reqCtx echo.Context) error {
	query := reqCtx.QueryParam("query")
	pageSizeStr := reqCtx.QueryParam("limit")
	reqOffsetStr := reqCtx.QueryParam("offset")
	limit, err := strconv.Atoi(pageSizeStr)

	if err != nil {
		limit = defaultSearchPageSize
	}

	offset, _ := strconv.Atoi(reqOffsetStr)

	searchResults, err := plugin.Search.Search(reqCtx.Request().Context(), query, offset, limit)

	if err != nil {
		return err
	}

	ctx := reqCtx.Request().Context()

	return views.SearchResults(searchResults).Render(ctx, reqCtx.Response().Writer)
}

func (plugin *SearchView) HeaderItem() core.Component {
	return core.ComponentFunc(views.Search().Render, 0)
}

func (plugin *SearchView) Shutdown(_ context.Context) error {
	return nil
}
