package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/internal/common"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

var errSqliteURLNotSet = errors.New("sqlite url is not set")

var _ core.Plugin = (*Repo)(nil)

const configKey = "$.core.repo.sqlite"

type Cfg struct {
	URL   string `yaml:"url"`
	Debug bool   `yaml:"debug"`
}

type Repo struct {
	Config *Cfg
	common.BaseBun
}

func (m *Repo) Setup(ctx context.Context, deps core.Dependencies) error {
	if m.Config == nil {
		m.Config = &Cfg{}

		err := deps.CfgReader().ReadAt(configKey, m.Config)
		if err != nil {
			return err
		}
	}

	sqldb, err := OpenSQLite(m.Config)
	if err != nil {
		return err
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	baseRepo, err := common.New(ctx, deps.Logger(), db, m.Config.Debug)
	m.BaseBun = baseRepo

	return err
}

func OpenSQLite(cfg *Cfg) (*sql.DB, error) {
	if cfg.URL == "" {
		return nil, errSqliteURLNotSet
	}

	sqldb, err := sql.Open(sqliteshim.ShimName, cfg.URL)
	if err != nil {
		return nil, err
	}

	if strings.Contains(cfg.URL, "memory") {
		const maxConnsInMemory = 1000

		sqldb.SetMaxIdleConns(maxConnsInMemory)
		sqldb.SetConnMaxLifetime(0)

		return sqldb, nil
	}

	sqldb.SetMaxOpenConns(1)
	sqldb.SetMaxIdleConns(0)

	return sqldb, nil
}

const updateDepsQuerySqlite = `DELETE FROM entity_mappings WHERE ref_entity_id IN (?0);
	INSERT INTO entity_mappings 
		(from_id, ref_entity_id, relation, to_id) 
	SELECT 
		entity.full_name, entity.full_name, rel.key, lookup.full_name
	FROM stored_entities as entity, 
		json_each(entity.relations) as rel, 
		json_each(rel.value) as reference
	JOIN 
		stored_entities AS lookup 
		ON lookup.full_name = reference.value 
		OR lookup.unkind_name = reference.value 
		OR lookup.name = reference.value
		OR lookup.full_name = REPLACE(reference.value, ':', ':default/')
	WHERE entity.full_name IN (?0);`

func (m *Repo) UpdateRelations(ctx context.Context, entityNames ...string) error {
	if len(entityNames) == 0 {
		return nil
	}

	return m.DB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.ExecContext(ctx, updateDepsQuerySqlite, bun.In(entityNames))
		return err
	})
}
