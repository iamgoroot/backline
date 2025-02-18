package indexer

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

type AsyncAPI struct {
}

func (plugin AsyncAPI) Setup(_ context.Context, _ core.Dependencies) error {
	return nil
}

func (plugin AsyncAPI) ProcessEntity(_ context.Context, _ core.Dependencies, _ *model.Entity) error {
	return nil
}
