package bluge

import (
	"context"
	"errors"
	"fmt"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	segment "github.com/blugelabs/bluge_segment_api"
)

func (plugin *Search) Index(_ context.Context, entityName, ref, category, value string) error {
	id := fmt.Sprintf("%s#%s", entityName, category)

	doc := bluge.NewDocument(id)

	doc.AddField(bluge.NewTextField("value", value).SearchTermPositions().StoreValue().HighlightMatches())
	doc.AddField(bluge.NewKeywordField("category", category).StoreValue())
	doc.AddField(bluge.NewKeywordField("entityName", entityName).StoreValue())
	doc.AddField(bluge.NewKeywordField("link", ref).StoreValue())

	return plugin.writer.Update(doc.ID(), doc)
}

func (plugin *Search) RemoveIndex(ctx context.Context, entityName ...string) error {
	reader, err := plugin.ensureReader()
	if err != nil {
		return err
	}

	for _, entityName := range entityName {
		err = errors.Join(err, plugin.removeEntity(ctx, reader, plugin.writer, entityName))
	}
	return err
}

func (plugin *Search) removeEntity(ctx context.Context, reader *bluge.Reader, writer *bluge.Writer, entityName string) error {
	q := bluge.NewTermQuery(entityName).SetField("entityName") //TODO: use multi term query
	request := bluge.NewAllMatches(q)

	var allErrs error

	_, err := executeRequest[struct{}](ctx, reader, request, func(match *search.DocumentMatch, _ *struct{}) segment.StoredFieldVisitor {
		return func(field string, value []byte) bool {
			if field == "_id" {
				id := string(value)
				deleteErr := writer.Delete(bluge.Identifier(id))
				allErrs = errors.Join(allErrs, deleteErr)
			}

			return true
		}
	})

	return errors.Join(allErrs, err)
}
