package bluge

import (
	"context"
	"fmt"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/highlight"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/iamgoroot/backline/pkg/core"
)

var errFailedOpeningBlugeReader = fmt.Errorf("error getting index reader")

func (plugin *Search) Search(ctx context.Context, query string, offset, limit int) ([]core.SearchResult, error) {
	reader, err := plugin.ensureReader()
	if err != nil {
		return nil, err
	}

	q := bluge.NewFuzzyQuery(query).SetField("value").SetFuzziness(1)
	request := bluge.NewTopNSearch(limit, q).SetFrom(offset).IncludeLocations()
	highlighter := highlight.NewHTMLHighlighter()

	return executeRequest[core.SearchResult](ctx, reader, request,
		func(match *search.DocumentMatch, result *core.SearchResult) segment.StoredFieldVisitor {
			return func(field string, value []byte) bool {
				switch field {
				case "link":
					result.Link = string(value)
				case "entityName":
					result.EntityName = string(value)
				case "category":
					result.Category = string(value)
				case "value":
					result.Highlight = highlighter.BestFragment(match.Locations["value"], value)
				}

				return true
			}
		})
}
