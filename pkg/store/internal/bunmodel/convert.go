package bunmodel

import (
	"github.com/iamgoroot/backline/pkg/model"
)

func StoredEntitiesToModels(storedEntities []*StoredEntity) []*model.Entity {
	models := make([]*model.Entity, len(storedEntities))
	for i, v := range storedEntities {
		models[i] = ToModel(v)
	}

	return models
}

func ConvertSpec(spec *model.Spec) Spec {
	return Spec{
		Type:       spec.Type,
		Lifecycle:  spec.Lifecycle,
		Owner:      spec.Owner,
		System:     spec.System,
		Definition: spec.Definition,
		Relations: Relations{
			OwnedBy:       spec.OwnedBy,
			OwnerOf:       spec.OwnerOf,
			ProvidesAPI:   spec.ProvidesAPI,
			ConsumesAPI:   spec.ConsumesAPI,
			APIConsumedBy: spec.APIConsumedBy,
			APIProvidedBy: spec.APIProvidedBy,
			DependsOn:     spec.DependsOn,
			DependencyOf:  spec.DependencyOf,
			ParentOf:      spec.ParentOf,
			ChildOf:       spec.ChildOf,
			MemberOf:      spec.MemberOf,
			HasMember:     spec.HasMember,
			PartOf:        spec.PartOf,
			HasPart:       spec.HasPart,
			Children:      spec.Children,
		},
	}
}

func convertSpecToModel(spec *Spec) model.Spec {
	return model.Spec{
		Type:       spec.Type,
		Lifecycle:  spec.Lifecycle,
		Owner:      spec.Owner,
		System:     spec.System,
		Definition: spec.Definition,
	}
}

func ToModel(entity *StoredEntity) *model.Entity {
	return &model.Entity{
		FullName:   entity.FullName,
		Metadata:   entity.Metadata,
		Spec:       getSpec(entity),
		APIVersion: entity.APIVersion,
		Kind:       entity.Kind,
	}
}
func getSpec(entity *StoredEntity) model.Spec {
	spec := convertSpecToModel(&entity.Spec)
	spec.Profile = &entity.Profile

	return spec
}
