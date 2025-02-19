package main

import (
	"log/slog"
	"os"

	"github.com/iamgoroot/backline/plugin/scanner"
	"github.com/iamgoroot/backline/plugin/search/bluge"
	"github.com/iamgoroot/backline/plugin/search/indexer"

	"github.com/iamgoroot/backline/app"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store"
	"github.com/iamgoroot/backline/plugin/discovery/github"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs"
)

func main() {
	// run only scanner
	application := app.App{

		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{ // Add Location Readers so backline knows how to read entities from different sources
				&github.Discovery{},
			},

			EntityRepo:        &store.Repo{},      // add repository plugin configurable with config file. supports pg and sqlite
			KeyValStore:       &store.KV{},        // add key-value storage plugin configurable with config file. supports pg and sqlite
			JobScheduler:      &store.Scheduler{}, // add job scheduler plugin configurable with config file. supports pg and sqlite
			DistributedLocker: &store.Locker{},    // add distributed lock plugin configurable with config file. supports pg and sqlite
			ScannerPlugin:     &scanner.Plugin{},  // add scanner plugin to scan/read entities
			SearchPlugin:      &bluge.Search{},
		},
		Plugins: []core.Plugin{
			&techdocs.Plugin{}, // add techdocs documentation plugin
			indexer.BaseEntityInfo{},
			&indexer.OpenAPIIndexer{},
		},
	}

	err := application.Run()

	if err != nil {
		slog.Error("Error while running. Shutting down", slog.Any("err", err))
		os.Exit(1)
	}
}
