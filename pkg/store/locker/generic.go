package locker

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/internal/common"
	"github.com/uptrace/bun"
)

const defaultLockExpires = time.Hour

// SQLLock can be used in cases when there is no other distributed lock support (like redis or pg).
type SQLLock struct {
	DB     *bun.DB
	logger core.Logger
}

type lock struct {
	Expires time.Time
	LockKey string `bun:",pk,notnull"`
}

func (m *SQLLock) Setup(ctx context.Context, deps core.Dependencies) error {
	m.logger = deps.Logger()
	_, err := m.DB.NewCreateTable().IfNotExists().Model((*lock)(nil)).Exec(ctx)

	return err
}

func (m *SQLLock) WithLock(ctx context.Context, lockKey string, run func(ctx context.Context) error) error {
	return common.EnsureSingleTx(ctx, m.DB, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}, func(ctx context.Context, tx bun.Tx) error {
		var notLocked bool

		var err error

		_, err = tx.NewDelete().Model((*lock)(nil)).Where("expires < ?", time.Now()).Exec(ctx)
		if err != nil {
			return err
		}

		lock := &lock{LockKey: lockKey, Expires: time.Now().Add(defaultLockExpires)}
		_, err = tx.NewInsert().Model(lock).Exec(ctx)

		defer func() {
			_, err = tx.NewDelete().Model(lock).WherePK().Exec(ctx)
			if err != nil {
				m.logger.Error("failed to release lock: ", slog.Any("err", err))
			}
		}()

		notLocked = err == nil

		if err != nil {
			return err
		}

		if notLocked {
			return run(ctx)
		}

		return core.ErrLockTaken("lock is already taken by another process")
	})
}

func (m *SQLLock) Shutdown(_ context.Context) error {
	return m.DB.Close()
}
