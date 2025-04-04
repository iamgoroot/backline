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
	"github.com/iamgoroot/backline/plugin/auth/oauth2"
	"github.com/iamgoroot/backline/plugin/catalog"
	"github.com/iamgoroot/backline/plugin/discovery/fs"
	"github.com/iamgoroot/backline/plugin/discovery/github"
	"github.com/iamgoroot/backline/plugin/documentation/asyncapi"
	"github.com/iamgoroot/backline/plugin/documentation/rawdefinition"
	"github.com/iamgoroot/backline/plugin/documentation/swaggerui"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs"
	"github.com/iamgoroot/backline/plugin/theme/stock"
)

func main() {
	application := app.App{
		// run catalog and scanner

		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{ // Add Location Readers so backline knows how to read entities from different sources
				&fs.Discovery{}, &github.Discovery{},
			},
			EntityRepo:        &store.Repo{}, // set repository plugin configurable with config file. supports pg and sqlite
			KeyValStore:       &store.KV{},   // set key-value storage plugin configurable with config file. supports pg and sqlite
			JobScheduler:      &store.Scheduler{},
			DistributedLocker: &store.Locker{},
			SearchPlugin:      &bluge.Search{},   // add search plugin (bluge uses local files for saving indices)
			ScannerPlugin:     &scanner.Plugin{}, // add scanner plugin to scan/read entities.
		},
		Plugins: []core.Plugin{
			catalog.Plugin{}, // add web interface for catalog
			oauth2.Plugin{},  // add oauth2 plugin for catalog authentication
			stock.Theme{},    // add stock theme for catalog

			&techdocs.Plugin{}, // add techdocs documentation plugin

			swaggerui.Plugin{},     // add ability to display openapi specs on entity pages
			asyncapi.Plugin{},      // add ability to display asyncapi specs on entity pages
			rawdefinition.Plugin{}, // add ability to display raw API definitions on entity pages

			indexer.BaseEntityInfo{},
			&indexer.OpenAPIIndexer{},
			&indexer.AsyncAPI{},
		},
	}

	err := application.Run()

	if err != nil {
		slog.Error("Error while running. Shutting down", slog.Any("err", err))
		os.Exit(1)
	}
}
