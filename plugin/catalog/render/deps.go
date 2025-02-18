package render

import (
	"context"

	"github.com/iamgoroot/backline/pkg/model"
)

type Repo interface {
	GetByName(ctx context.Context, name string) (model.Entity, error)
	List(ctx context.Context, params *model.ListEntityReq) ([]model.Entity, error)
}
