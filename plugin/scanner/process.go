package scanner

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

var (
	errDefinitionDownloadFailed = errors.New("failed to download definition")
	errInvalidEntity            = errors.New("invalid entity or not an entity")
	errKindLocationNotSupported = errors.New("location kind is not supported. use properly configured scanner instead")
)

var preProcessors = []EntityProcessor{normalizeNames{}, populateRelations{}, definitionDownloader{}}

type normalizeNames struct{}

func (e normalizeNames) ProcessEntity(_ context.Context, _ core.Dependencies, entity *model.Entity) error {
	if entity.Metadata.Name == "" {
		return fmt.Errorf("%w: %s downloaded by %s", errInvalidEntity, entity.LocationMetadata.Location, entity.LocationMetadata.DownloadedBy)
	}

	if entity.Kind == "Location" {
		return errKindLocationNotSupported
	}

	entity.Kind = strings.ToLower(entity.Kind)
	if entity.Metadata.Namespace == "" {
		entity.Metadata.Namespace = "default"
	}

	entity.FullName = fmt.Sprintf("%s:%s/%s", strings.ToLower(entity.Kind), entity.Metadata.Namespace, entity.Metadata.Name)

	return nil
}

type populateRelations struct{}

func (e populateRelations) ProcessEntity(_ context.Context, _ core.Dependencies, entity *model.Entity) error {
	if entity.Spec.System != "" {
		entity.Spec.PartOf = append(entity.Spec.PartOf, entity.Spec.System)
	}

	if entity.Spec.Owner != "" {
		entity.Spec.OwnedBy = append(entity.Spec.OwnedBy, entity.Spec.Owner)
	}

	if entity.Spec.Parent != "" {
		entity.Spec.ChildOf = append(entity.Spec.ChildOf, entity.Spec.Parent)
	}

	return nil
}

type definitionDownloader struct{}

func (e definitionDownloader) ProcessEntity(ctx context.Context, deps core.Dependencies, entity *model.Entity) error {
	if entity.Spec.Definition == nil {
		return nil
	}

	txt := getDefinitionTxt(ctx, deps, entity)
	if txt == "" {
		return errDefinitionDownloadFailed
	}

	entity.Spec.Definition = txt

	return nil
}

func getDefinitionTxt(ctx context.Context, deps core.Dependencies, entity *model.Entity) string {
	def := entity.Spec.Definition

	val, ok := def.(string)
	if ok {
		return val
	}

	ref := parseDefinitionRef(def)
	if ref == "" {
		return ""
	}

	for _, d := range deps.Discoveries() {
		data, err := d.TryDownload(ctx, deps, entity.LocationMetadata, ref)
		if err == nil && data != "" {
			return data
		}
	}

	return ""
}

func parseDefinitionRef(def any) string {
	typedMap, ok := def.(map[string]any)
	if !ok {
		return ""
	}

	var result string

	safeMap(typedMap, "$text", func(ref string) { // don't like this part of backstage logic. TODO: stitch that on read?
		result = ref
	})
	safeMap(typedMap, "$openapi", func(ref string) {
		result = ref
	})
	safeMap(typedMap, "$json", func(ref string) {
		result = ref
	})
	safeMap(typedMap, "$yaml", func(ref string) {
		result = ref
	})

	return result
}
func safeMap[T any](mp map[string]any, key string, run func(val T)) {
	if val, ok := mp[key]; ok {
		if t, ok := val.(T); ok {
			run(t)
		}
	}
}
