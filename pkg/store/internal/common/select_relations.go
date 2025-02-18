package common

import (
	"context"
	"strings"

	"github.com/iamgoroot/backline/pkg/model"
	"github.com/uptrace/bun/dialect"
)

// TODO: move into respective implementations

func (m baseRepo) populateDirectRelations(ctx context.Context, entityID string, entity *model.Entity) error {
	var query string

	switch m.db.Dialect().Name() {
	case dialect.PG:
		query = `SELECT	relation, ARRAY_AGG(to_id) AS related FROM entity_mappings WHERE from_id = ? GROUP BY relation;`
	case dialect.SQLite:
		query = `SELECT	relation, GROUP_CONCAT(to_id) AS related FROM entity_mappings WHERE from_id = ? GROUP BY relation;`
	}

	return m.ProcessRelations(ctx, query, entityID, entity, directRelation)
}

func (m baseRepo) populateReverseRelations(ctx context.Context, entityID string, entity *model.Entity) error {
	var query string

	switch m.db.Dialect().Name() {
	case dialect.PG:
		query = `SELECT	relation, ARRAY_AGG(from_id) AS related FROM entity_mappings WHERE to_id = ? GROUP BY relation;`
	case dialect.SQLite:
		query = `SELECT	relation, GROUP_CONCAT(from_id) AS related FROM entity_mappings WHERE to_id = ? GROUP BY relation;`
	}

	return m.ProcessRelations(ctx, query, entityID, entity, reverseRelation)
}

// ProcessRelations processes relations and reversed relations
// nolint goconst more readable this way
func (m baseRepo) ProcessRelations(ctx context.Context, query, entityID string, entity *model.Entity, relationResolver func(string) string) error {
	rows, err := m.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var relation, relatedStr string

	for rows.Next() {
		err := rows.Scan(&relation, &relatedStr)
		if err != nil {
			return err
		}
		relatedStr = strings.TrimSuffix(relatedStr, "}")
		relatedStr = strings.TrimPrefix(relatedStr, "{")
		related := strings.Split(relatedStr, ",")
		switch relationResolver(relation) {
		case "ParentOf":
			entity.Spec.ParentOf = related
		case "ChildOf":
			entity.Spec.ChildOf = related
		case "Children":
			entity.Spec.Children = related
		case "OwnedBy":
			entity.Spec.OwnedBy = related
		case "OwnerOf":
			entity.Spec.OwnerOf = related
		case "DependsOn":
			entity.Spec.DependsOn = related
		case "DependencyOf":
			entity.Spec.DependencyOf = related
		case "ProvidesAPI":
			entity.Spec.ProvidesAPI = related
		case "APIProvidedBy":
			entity.Spec.APIProvidedBy = related
		case "ConsumesAPI":
			entity.Spec.ConsumesAPI = related
		case "APIConsumedBy":
			entity.Spec.APIConsumedBy = related
		case "MemberOf":
			entity.Spec.MemberOf = related
		case "HasMember":
			entity.Spec.HasMember = related
		case "PartOf":
			entity.Spec.PartOf = related
		case "HasPart":
			entity.Spec.HasPart = related
		}
	}
	return nil
}

func directRelation(relation string) string {
	return relation
}
func reverseRelation(relation string) string { // nolint cyclop
	switch relation {
	case "ParentOf":
		return "ChildOf"
	case "ChildOf":
		return "ParentOf"
	case "Children":
		return "ParentOf"
	case "OwnedBy":
		return "OwnerOf"
	case "OwnerOf":
		return "OwnedBy"
	case "DependsOn":
		return "DependencyOf"
	case "DependencyOf":
		return "DependsOn"
	case "ProvidesAPI":
		return "APIProvidedBy"
	case "APIProvidedBy":
		return "ProvidesAPI"
	case "ConsumesAPI":
		return "APIConsumedBy"
	case "APIConsumedBy":
		return "ConsumesAPI"
	case "MemberOf":
		return "HasMember"
	case "HasMember":
		return "MemberOf"
	case "PartOf":
		return "HasPart"
	case "HasPart":
		return "PartOf"
	default:
		return ""
	}
}
