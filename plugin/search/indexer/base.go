package indexer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

type BaseEntityInfo struct {
}

func (plugin BaseEntityInfo) Shutdown(_ context.Context) error {
	return nil
}

func (plugin BaseEntityInfo) Setup(_ context.Context, _ core.Dependencies) error {
	return nil
}

func (plugin BaseEntityInfo) ProcessEntity(ctx context.Context, deps core.Dependencies, entity *model.Entity) error {
	indexer := deps.Search()
	if indexer == nil {
		return nil
	}

	link := fmt.Sprintf("/catalog/view/entity/fullname/%s", entity.FullName)

	return errors.Join(
		indexer.Index(ctx, entity.FullName, link, "title", entity.Metadata.Title),
		indexer.Index(ctx, entity.FullName, link, "name", entity.Metadata.Name),
		indexer.Index(ctx, entity.FullName, link, "tags", strings.Join(entity.Metadata.Tags, "\n")),
	)
}
