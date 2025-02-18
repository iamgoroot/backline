package bluge

import (
	"context"
	"errors"
	"fmt"
	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search/highlight"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/search/ui"
	"log"
)

type Search struct {
	ui.SearchView
	cfg bluge.Config
}

func (plugin *Search) Setup(ctx context.Context, deps core.Dependencies) error {
	plugin.cfg = bluge.DefaultConfig("./bluge-index")
	return plugin.SearchView.Setup(ctx, deps)
}

func (plugin *Search) Shutdown(_ context.Context) error {
	return nil
}

func (plugin *Search) Search(_ context.Context, query string, offset, limit int) ([]core.SearchResult, error) {
	reader, err := bluge.OpenReader(plugin.cfg)
	if err != nil {
		log.Fatalf("error getting index reader: %v", err)
	}

	defer reader.Close()

	q := bluge.NewFuzzyQuery(query).SetField("value").SetFuzziness(1)
	request := bluge.NewTopNSearch(limit, q).SetFrom(offset).IncludeLocations()
	documentMatchIterator, err := reader.Search(context.Background(), request)
	if err != nil {
		return nil, err
	}
	highligher := highlight.NewHTMLHighlighter()
	var searchResults []core.SearchResult

	match, err := documentMatchIterator.Next()

	if err != nil {
		return nil, err
	}

	for match != nil {
		result := core.SearchResult{}
		err = match.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case "link":
				result.Link = string(value)
			case "entityName":
				result.EntityName = string(value)
			case "category":
				result.Category = string(value)
			case "value":
				result.Highlight = highligher.BestFragment(match.Locations["value"], value)
			}
			return true
		})
		if err != nil {
			return nil, err
		}

		searchResults = append(searchResults, result)

		match, err = documentMatchIterator.Next()

		if err != nil {
			return nil, err
		}
	}

	return searchResults, nil
}

func (plugin *Search) Index(_ context.Context, entityName, ref, category, value string) error {
	writer, err := bluge.OpenWriter(plugin.cfg)

	if err != nil {
		return err
	}

	defer writer.Close()

	id := fmt.Sprintf("%s#%s", entityName, category)

	doc := bluge.NewDocument(id)

	doc.AddField(bluge.NewTextField("value", value).SearchTermPositions().StoreValue().HighlightMatches())
	doc.AddField(bluge.NewKeywordField("category", category).StoreValue())
	doc.AddField(bluge.NewKeywordField("entityName", entityName).StoreValue())
	doc.AddField(bluge.NewKeywordField("link", ref).StoreValue())

	return writer.Update(doc.ID(), doc)
}
func (plugin *Search) RemoveIndex(ctx context.Context, entityName ...string) error {
	var err error
	for _, entityName := range entityName {
		err = errors.Join(err, plugin.getIDs(ctx, entityName)) //TODO: reuse writer/reader
	}
	return err
}
func (plugin *Search) getIDs(ctx context.Context, entityName string) error {
	writer, err := bluge.OpenWriter(plugin.cfg)

	if err != nil {
		return err
	}

	defer writer.Close()
	reader, err := writer.Reader()
	q := bluge.NewTermQuery(entityName).SetField("entityName") //use multi term query
	request := bluge.NewAllMatches(q)

	documentMatchIterator, err := reader.Search(context.Background(), request)
	if err != nil {
		return err
	}
	match, err := documentMatchIterator.Next()

	if err != nil {
		return err
	}
	var allErrs error

	for match != nil {
		err = match.VisitStoredFields(func(field string, value []byte) bool {
			if field == "_id" {
				id := string(value)
				deleteErr := writer.Delete(bluge.Identifier(id))
				allErrs = errors.Join(allErrs, deleteErr)
			}
			return true
		})

		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}

		match, err = documentMatchIterator.Next()

		if err != nil {
			allErrs = errors.Join(allErrs, err)
		}
	}

	return allErrs
}
