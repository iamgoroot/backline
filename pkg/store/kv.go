package store

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/kv"
)

type KV struct {
	core.StoreKV
}

const kvCfgPath = "$.core.kv"

type kvCfg repoCfg

func (m *KV) Setup(ctx context.Context, deps core.Dependencies) error {
	cfg := &kvCfg{}

	err := deps.CfgReader().ReadAt(kvCfgPath, cfg)
	if err != nil {
		return err
	}

	m.StoreKV = resolveKV(cfg)
	if m.StoreKV == nil {
		return core.ConfigurationError(kvCfgPath)
	}

	return m.StoreKV.Setup(ctx, deps)
}

func resolveKV(cfg *kvCfg) core.StoreKV {
	switch {
	case cfg.PG != nil:
		return &kv.PgKV{Config: cfg.PG}
	case cfg.Sqlite != nil:
		return &kv.SqliteKV{Config: cfg.Sqlite}
	}

	return nil
}
