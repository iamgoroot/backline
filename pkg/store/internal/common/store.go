package common

import (
	"context"
	"fmt"

	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/pkg/store/internal/bunmodel"
	"github.com/uptrace/bun"
)

func (m baseRepo) Store(ctx context.Context, entityModel *model.Entity) error {
	return m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		unkindName := fmt.Sprintf("%s/%s", entityModel.Metadata.Namespace, entityModel.Metadata.Name)
		fullName := fmt.Sprintf("%s:%s", entityModel.Kind, unkindName)

		entity := bunmodel.StoredEntity{
			FullName:   fullName,
			UnkindName: unkindName,
			APIVersion: entityModel.APIVersion,
			Kind:       entityModel.Kind,
			Metadata:   entityModel.Metadata,
			Spec:       bunmodel.ConvertSpec(&entityModel.Spec),
		}
		if entityModel.Spec.Profile != nil {
			entity.Profile = *entityModel.Spec.Profile
		}

		_, err := tx.NewInsert().Model(&entity).On("CONFLICT (full_name) DO UPDATE").Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
