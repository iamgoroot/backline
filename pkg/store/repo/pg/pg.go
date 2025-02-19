package pg

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store/internal/common"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Cfg struct {
	TLSConfig *tls.Config
	DSN       string `yaml:"dsn"`
	Host      string `yaml:"host"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	Port      int    `yaml:"port"`
	Debug     bool   `yaml:"debug"`
	Insecure  bool   `yaml:"insecure"`
}

const configKey = "$.core.repo.pg"

type Repo struct {
	Config *Cfg
	DB     *bun.DB
	common.BaseBun
}

func (m *Repo) Setup(ctx context.Context, deps core.Dependencies) error {
	db, err := m.getDB(deps)
	if err != nil {
		return err
	}

	baseRepo, err := common.New(ctx, deps.Logger(), db, m.Config.Debug)
	m.BaseBun = baseRepo
	m.DB = db

	return err
}

func (m *Repo) getDB(deps core.Dependencies) (*bun.DB, error) {
	if m.Config == nil {
		m.Config = &Cfg{}

		err := deps.CfgReader().ReadAt(configKey, m.Config)
		if err != nil {
			return nil, err
		}
	}

	if m.DB != nil {
		return m.DB, nil
	}

	sqldb := OpenPG(m.Config)
	db := bun.NewDB(sqldb, pgdialect.New())

	return db, nil
}

func OpenPG(cfg *Cfg) *sql.DB {
	var driverCfg *pgdriver.Connector
	if cfg.DSN != "" {
		driverCfg = pgdriver.NewConnector(
			pgdriver.WithDSN(cfg.DSN),
			pgdriver.WithTLSConfig(cfg.TLSConfig),
			pgdriver.WithApplicationName("backline"),
		)
	} else {
		driverCfg = pgdriver.NewConnector(
			pgdriver.WithAddr(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
			pgdriver.WithDatabase(cfg.Database),
			pgdriver.WithUser(cfg.User),
			pgdriver.WithPassword(cfg.Password),
			pgdriver.WithTLSConfig(cfg.TLSConfig),
			pgdriver.WithApplicationName("backline"),
		)
	}

	return sql.OpenDB(driverCfg)
}

const updateDepsPg = `DELETE FROM entity_mappings WHERE ref_entity_id IN (?0);
INSERT INTO entity_mappings 
    (from_id, ref_entity_id, relation, to_id)
SELECT 
    entity.full_name,
    entity.full_name, 
    rel.key, 
    lookup.full_name
FROM 
    stored_entities AS entity,
    LATERAL jsonb_each(entity.relations) AS rel(key, value),
    LATERAL jsonb_array_elements_text(rel.value) AS reference
JOIN 
    stored_entities AS lookup 
    ON lookup.full_name = reference 
       OR lookup.unkind_name = reference 
       OR lookup.name = reference 
       OR lookup.full_name = REPLACE(reference, ':', ':default/')
WHERE entity.full_name IN (?0);`

func (m *Repo) UpdateRelations(ctx context.Context, entityNames ...string) error {
	if len(entityNames) == 0 {
		return nil
	}

	return m.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.ExecContext(ctx, updateDepsPg, bun.In(entityNames))
		return err
	})
}
