package kv

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type bunKV struct {
	db     *bun.DB
	logger core.Logger
}

type keyVal struct {
	ExpiresAt bun.NullTime
	UpdatedAt time.Time
	Group     string          `bun:",unique:key"`
	Key       string          `bun:",unique:key"`
	Value     json.RawMessage `bun:"type:jsonb"`
}

func (m bunKV) setup(ctx context.Context) error {
	_, err := m.db.NewCreateTable().IfNotExists().Model((*keyVal)(nil)).Exec(ctx)
	if err != nil {
		return err
	}

	go func() {
		tick := time.Tick(time.Hour)
		for range tick {
			cleanUpErr := m.keyValCleanup(ctx)
			if cleanUpErr != nil {
				m.logger.Error("failed to cleanup keyvals: ", slog.Any("err", cleanUpErr))
			}
		}
	}()

	return err
}

func (m bunKV) Set(ctx context.Context, keyGroup, key string, value any, ttl time.Duration) error {
	var expire time.Time
	if ttl > 0 {
		expire = time.Now().Add(ttl)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	record := &keyVal{
		Group:     keyGroup,
		Key:       key,
		Value:     data,
		ExpiresAt: schema.NullTime{Time: expire},
		UpdatedAt: time.Now(),
	}
	_, err = m.db.NewInsert().Model(record).
		On(`CONFLICT ("group", key) 
			DO UPDATE SET 
				value = EXCLUDED.value, 
				expires_at = EXCLUDED.expires_at,  
				updated_at = EXCLUDED.updated_at`).
		Exec(ctx)

	return err
}

func (m bunKV) Get(ctx context.Context, keyGroup, key string, value any) error {
	var bunKV keyVal
	err := m.db.NewSelect().
		Model(&bunKV).
		Where(`"group" = ?`, keyGroup).
		Where("key = ?", key).
		WhereGroup("AND", func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq.Where("expires_at IS NULL")
			sq.WhereOr("expires_at > ?", time.Now())
			return sq
		}).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.NotFoundError(err.Error())
		}

		return err
	}

	err = json.Unmarshal(bunKV.Value, &value)

	return err
}

func (m bunKV) keyValCleanup(ctx context.Context) error {
	_, err := m.db.NewDelete().Model((*keyVal)(nil)).
		Where("expires_at IS NOT NULL").
		Where("expires_at < ?", time.Now()).Exec(ctx)

	return err
}

func (m bunKV) Shutdown(_ context.Context) error {
	return m.db.Close()
}
