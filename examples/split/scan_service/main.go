package main

import (
	"log/slog"
	"os"

	"github.com/iamgoroot/backline/pkg/store/kv"
	"github.com/iamgoroot/backline/pkg/store/repo/sqlite"
	"github.com/iamgoroot/backline/plugin/scanner"
	"github.com/iamgoroot/backline/plugin/search/bluge"
	"github.com/iamgoroot/backline/plugin/search/indexer"

	"github.com/iamgoroot/backline/app"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/discovery/github"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs"
)

func main() {
	// run only scanner
	application := app.App{

		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{ // Add Location Readers so backline knows how to read entities from different sources
				&github.Discovery{}, // read entities from github
			},

			EntityRepo:    &sqlite.Repo{},    // add repository plugin configurable with config file. supports pg and sqlite
			KeyValStore:   &kv.SqliteKV{},    // add key-value storage plugin configurable with config file. supports pg and sqlite
			ScannerPlugin: &scanner.Plugin{}, // add scanner plugin to scan/read entities
			SearchPlugin:  &bluge.Search{},   // add search plugin to allow entity search
		},
		Plugins: []core.Plugin{
			&techdocs.Plugin{}, // add techdocs documentation plugin

			indexer.BaseEntityInfo{},  // indexes common entity info while scanning
			&indexer.OpenAPIIndexer{}, // indexes openapi specs
			&indexer.AsyncAPI{},       // indexes asyncapi specs
		},
	}

	err := application.Run()

	if err != nil {
		slog.Error("Error while running. Shutting down", slog.Any("err", err))
		os.Exit(1)
	}
}
