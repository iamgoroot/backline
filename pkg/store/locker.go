package store

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/locker"
	"github.com/iamgoroot/backline/pkg/store/repo/sqlite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

type Locker struct {
	core.DistributedLock
}

const lockCfgPath = "$.core.lock"

type lockCfg repoCfg

func (m *Locker) Setup(_ context.Context, deps core.Dependencies) error {
	cfg := &lockCfg{}

	err := deps.CfgReader().ReadAt(lockCfgPath, cfg)
	if err != nil {
		deps.Logger().Error("failed to read locker configuration", "err", err)
	}

	m.DistributedLock = resolveLocker(deps.Logger(), cfg)
	if m.DistributedLock == nil {
		return core.ConfigurationError(lockCfgPath)
	}

	return nil
}

func resolveLocker(logger core.Logger, cfg *lockCfg) core.DistributedLock {
	switch {
	case cfg.PG != nil:
		return &locker.PgLock{Config: cfg.PG}
	case cfg.Sqlite != nil:
		sqldb, err := sqlite.OpenSQLite(cfg.Sqlite)
		if err != nil {
			return nil
		}

		db := bun.NewDB(sqldb, sqlitedialect.New())

		return &locker.SQLLock{DB: db}
	default:
		logger.Warn("no locker configuration found. Using NoOP locker. Please configure locker if you run multiple instances.")
		return &locker.NoOP{}
	}
}
