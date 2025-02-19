package common

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/pkg/store/internal/bunmodel"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
)

type BaseBun interface {
	Shutdown(context.Context) error
	List(context.Context, *model.ListEntityReq) ([]*model.Entity, error)
	GetByName(context.Context, string) (*model.Entity, error)
	Store(context.Context, *model.Entity) error
	UpdateAsOrphans(ctx context.Context, at *time.Time, entityNames ...string) error
	DB() *bun.DB
	Logger() core.Logger
}

func New(ctx context.Context, logger core.Logger, db *bun.DB, debug bool) (BaseBun, error) {
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(debug), bundebug.WithVerbose(debug)))
	repo := baseRepo{
		logger: logger,
		db:     db,
	}

	return repo, repo.setup(ctx)
}

type baseRepo struct {
	logger core.Logger
	db     *bun.DB
}

func (m baseRepo) setup(ctx context.Context) error {
	m.db.RegisterModel((*bunmodel.EntityMapping)(nil))
	entities := []any{(*bunmodel.StoredEntity)(nil), (*bunmodel.EntityMapping)(nil)}

	for _, entity := range entities {
		_, err := m.db.NewCreateTable().IfNotExists().Model(entity).Exec(ctx)
		if err != nil {
			return err
		}
	}

	_, err := m.db.NewCreateIndex().
		Model((*bunmodel.StoredEntity)(nil)).
		IfNotExists().
		Index("entity_unkind_name_idx").
		Column("unkind_name").
		Exec(ctx)

	if err != nil {
		return err
	}

	_, err = m.db.NewCreateIndex().Model((*bunmodel.EntityMapping)(nil)).IfNotExists().Index("from_id_idx").Column("from_id").Exec(ctx)

	return err
}

var allowedSorts = []string{"kind", "name"}

func (m baseRepo) List(ctx context.Context, params *model.ListEntityReq) ([]*model.Entity, error) {
	var storedEntities []*bunmodel.StoredEntity

	query := m.db.NewSelect().Model(&storedEntities)
	if params.Kind != "" {
		query.Where("kind = ?", strings.ToLower(params.Kind))
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	if !params.ShowOrphans {
		query.Where("orphaned_at IS NULL")
	}

	query.Limit(params.Limit)
	query.Offset(params.Offset)

	if params.Sort != "" {
		var sort, order = params.Sort, "ASC"
		if strings.HasPrefix(params.Sort, "-") {
			order = "DESC"
			sort = params.Sort[1:]
		} else if strings.HasPrefix(params.Sort, "+") {
			sort = params.Sort[1:]
		}

		if slices.Contains(allowedSorts, sort) {
			query.Order(fmt.Sprintf("%s %s", sort, order))
		}
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return bunmodel.StoredEntitiesToModels(storedEntities), nil
}

func (m baseRepo) GetByName(ctx context.Context, name string) (*model.Entity, error) {
	entity := &bunmodel.StoredEntity{}
	kind, namespace, shortname := model.ParseFullName(name)

	err := m.db.NewSelect().
		Model(entity).
		Where("full_name=?", name).
		WhereGroup("OR", func(subQuery *bun.SelectQuery) *bun.SelectQuery {
			if kind != "" {
				subQuery.Where("kind = ?", strings.ToLower(kind))
			}
			if namespace != "" {
				subQuery.Where("namespace = ?", namespace)
			}
			if shortname != "" {
				subQuery.Where("name = ?", shortname)
			}
			return subQuery
		}).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return &model.Entity{}, err
	}

	entityModel := bunmodel.ToModel(entity)

	err = m.populateDirectRelations(ctx, entity.FullName, entityModel)
	if err != nil {
		return &model.Entity{}, err
	}

	err = m.populateReverseRelations(ctx, entity.FullName, entityModel)
	if err != nil {
		return &model.Entity{}, err
	}

	return entityModel, err
}

func (m baseRepo) UpdateAsOrphans(ctx context.Context, orphanedAt *time.Time, entityNames ...string) error {
	if len(entityNames) == 0 {
		return nil
	}

	return m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().Model((*bunmodel.StoredEntity)(nil)).
			Where("full_name IN (?)", bun.In(entityNames)).
			Set("orphaned_at = ?", orphanedAt).
			Exec(ctx)

		return err
	})
}

func (m baseRepo) Shutdown(_ context.Context) error {
	return m.db.Close()
}
func (m baseRepo) DB() *bun.DB {
	return m.db
}
func (m baseRepo) Logger() core.Logger {
	return m.logger
}
