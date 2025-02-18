package elastic

import (
	"context"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/search/ui"
)

const defaultSearchIndexName = "backline-search"

type Cfg struct {
	IndexName              string   `yaml:"indexName"`
	Username               string   `yaml:"username"`
	Password               string   `yaml:"password"`
	CloudID                string   `yaml:"cloudId"`
	APIKey                 string   `yaml:"apiKey"`
	ServiceToken           string   `yaml:"serviceToken"`
	CertificateFingerprint string   `yaml:"certificateFingerprint"`
	Addresses              []string `yaml:"addresses"`
	NumberOfShards         int      `yaml:"numberOfShards"`
	NumberOfReplicas       int      `yaml:"numberOfReplicas"`
}
type Search struct {
	ui.SearchView
	logger core.Logger
	client *elasticsearch.Client
	cfg    Cfg
}

func (plugin *Search) Setup(ctx context.Context, deps core.Dependencies) error {
	err := deps.CfgReader().ReadAt("$.search.elastic", &plugin.cfg)
	if err != nil {
		return err
	}

	elasticCfg := elasticsearch.Config{
		Addresses:              plugin.cfg.Addresses,
		Username:               plugin.cfg.Username,
		Password:               plugin.cfg.Password,
		APIKey:                 plugin.cfg.APIKey,
		CloudID:                plugin.cfg.CloudID,
		ServiceToken:           plugin.cfg.ServiceToken,
		CertificateFingerprint: plugin.cfg.CertificateFingerprint,
	}

	if plugin.cfg.NumberOfReplicas == 0 {
		plugin.cfg.NumberOfReplicas = 1
	}

	if plugin.cfg.NumberOfShards == 0 {
		plugin.cfg.NumberOfShards = 1
	}

	if plugin.cfg.IndexName == "" {
		plugin.cfg.IndexName = defaultSearchIndexName
	}

	esClient, err := elasticsearch.NewClient(elasticCfg)

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	plugin.client = esClient
	plugin.logger = deps.Logger()
	err = plugin.createIndex(ctx)

	if err != nil {
		return err
	}

	return plugin.SearchView.Setup(ctx, deps)
}

func (plugin *Search) Shutdown(_ context.Context) error {
	return nil
}
