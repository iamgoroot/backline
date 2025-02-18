package locker

import (
	"context"
	"database/sql"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/repo/pg"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// PgLock Use when there is no other distributed lock support (like redis).
type PgLock struct {
	Config *pg.Cfg
	DB     *bun.DB
}

const pgConfigKey = "$.core.lock.pg"

func (m *PgLock) Setup(_ context.Context, deps core.Dependencies) error {
	var err error
	m.DB, err = m.getDB(deps)

	return err
}
func (m *PgLock) getDB(deps core.Dependencies) (*bun.DB, error) {
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

func (m *PgLock) WithLock(ctx context.Context, lockKey string, run func(ctx context.Context) error) error {
	return m.DB.RunInTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}, func(ctx context.Context, tx bun.Tx) error {
		var notLocked bool

		err := tx.NewRaw("SELECT pg_try_advisory_xact_lock(hashtext(?))", lockKey).Scan(ctx, &notLocked)
		if err != nil {
			return err
		}

		if notLocked {
			return run(ctx)
		}

		return core.ErrLockTaken("lock is already taken by another process")
	})
}

func (m *PgLock) Shutdown(_ context.Context) error {
	return m.DB.Close()
}
