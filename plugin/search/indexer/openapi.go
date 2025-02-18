package indexer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

type OpenAPIIndexer struct {
	core.NoOpShutdown
	LinkPattern string
}

func (plugin *OpenAPIIndexer) Setup(_ context.Context, _ core.Dependencies) error {
	if plugin.LinkPattern == "" {
		plugin.LinkPattern = "/swagger-ui/view/by-fullname/%s"
	}

	return nil
}

func (plugin *OpenAPIIndexer) ProcessEntity(ctx context.Context, deps core.Dependencies, entity *model.Entity) error {
	if entity.Spec.Type != "openapi" {
		return nil
	}

	var definition string

	var ok bool
	if definition, ok = entity.Spec.Definition.(string); !ok {
		deps.Logger().Debug("cannot index definition of entity", slog.String("entity", entity.FullName))
	}

	spec, err := parseYAMLOrJSON([]byte(definition))
	if err != nil {
		return err
	}

	link := fmt.Sprintf(plugin.LinkPattern, entity.FullName)
	indexFunc := func(path []string, val string) error {
		indexName := strings.Join(path, ".")
		return deps.Search().Index(ctx, entity.FullName, link, indexName, val)
	}

	err = errors.Join(
		handleSpec(spec, indexFunc, "info", "title"),
		handleSpec(spec, indexFunc, "info", "description"),
		handleSpec(spec, indexFunc, "paths", "*", "description"),
	)
	if err != nil {
		deps.Logger().Info("error while indexing openapi definition of entity",
			slog.String("entity", entity.FullName),
			slog.String("error", err.Error()))
	}

	return nil
}
