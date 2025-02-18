package store

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/repo/pg"
	"github.com/iamgoroot/backline/pkg/store/repo/sqlite"
)

const repoCfgPath = "$.core.repo"

type Repo struct {
	core.Repo
}

type repoCfg struct {
	PG     *pg.Cfg     `yaml:"pg,omitempty"`
	Sqlite *sqlite.Cfg `yaml:"sqlite,omitempty"`
}

func (m *Repo) Setup(ctx context.Context, deps core.Dependencies) error {
	cfg := &repoCfg{}

	err := deps.CfgReader().ReadAt(repoCfgPath, cfg)
	if err != nil {
		return err
	}

	m.Repo = resolveRepo(cfg)
	if m.Repo == nil {
		return core.ConfigurationError(repoCfgPath)
	}

	return m.Repo.Setup(ctx, deps)
}

func resolveRepo(cfg *repoCfg) core.Repo {
	switch {
	case cfg.PG != nil:
		return &pg.Repo{Config: cfg.PG}
	case cfg.Sqlite != nil:
		return &sqlite.Repo{Config: cfg.Sqlite}
	}

	return nil
}
