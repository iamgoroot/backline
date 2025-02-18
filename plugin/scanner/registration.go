package scanner

import (
	"context"
	"log/slog"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

type entityRegistrator struct {
	ctx             context.Context // TODO: remove context from here
	deps            core.Dependencies
	registeredNames map[string]struct{}
}

func (r *entityRegistrator) RegisterEntity(locationMeta *model.LocationMetadata, entity *model.Entity) {
	entity.LocationMetadata = locationMeta
	err := r.processEntity(entity, preProcessors)

	if err != nil {
		r.deps.Logger().Error("failed to process entity", slog.Any("err", err), slog.String("entity", entity.Metadata.Name))
		return
	}

	err = r.deps.Plugins().ProcessEntity(r.ctx, r.deps, entity)
	if err != nil {
		r.deps.Logger().Error("failed to process entity", slog.Any("err", err), slog.String("entity", entity.Metadata.Name))
		return
	}

	err = r.deps.Repo().Store(r.ctx, entity)
	if err != nil {
		r.deps.Logger().Error("failed to store entity", slog.Any("err", err), slog.String("entity", entity.Metadata.Name))
		return
	}

	r.registeredNames[entity.FullName] = struct{}{}
}

func (r *entityRegistrator) processEntity(entity *model.Entity, processors []EntityProcessor) error {
	for _, p := range processors {
		if err := p.ProcessEntity(r.ctx, r.deps, entity); err != nil {
			return err
		}
	}

	return nil
}
