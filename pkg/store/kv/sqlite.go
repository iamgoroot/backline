package kv

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/repo/sqlite"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

const sqliteConfigKey = "$.core.kv.sqlite"

type SqliteKV struct {
	Config *sqlite.Cfg
	bunKV
}

func (m *SqliteKV) Setup(ctx context.Context, deps core.Dependencies) error {
	if m.Config == nil {
		m.Config = &sqlite.Cfg{}

		err := deps.CfgReader().ReadAt(sqliteConfigKey, m.Config)
		if err != nil {
			return err
		}
	}

	sqldb, err := sqlite.OpenSQLite(m.Config)
	if err != nil {
		return err
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	m.bunKV = bunKV{db: db, logger: deps.Logger()}

	return m.bunKV.setup(ctx)
}

func (m *SqliteKV) Shutdown(ctx context.Context) error {
	return m.bunKV.Shutdown(ctx)
}
