package kv

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/repo/pg"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type PgKV struct {
	DB *bun.DB
	bunKV
	Config *pg.Cfg
}

const pgConfigKey = "$.core.kv.pg"

func (m *PgKV) Setup(ctx context.Context, deps core.Dependencies) error {
	db, err := m.getDB(deps)
	if err != nil {
		return err
	}

	m.DB = db
	m.bunKV = bunKV{db: db, logger: deps.Logger()}

	return m.bunKV.setup(ctx)
}

func (m *PgKV) getDB(deps core.Dependencies) (*bun.DB, error) {
	if m.Config == nil {
		m.Config = &pg.Cfg{}

		err := deps.CfgReader().ReadAt(pgConfigKey, m.Config)
		if err != nil {
			return nil, err
		}
	}

	if m.DB != nil {
		return m.DB, nil
	}

	sqldb := pg.OpenPG(m.Config)
	db := bun.NewDB(sqldb, pgdialect.New())

	return db, nil
}
