package main

import (
	"log/slog"
	"os"

	"github.com/iamgoroot/backline/plugin/search/elastic"

	"github.com/iamgoroot/backline/plugin/scanner"

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
				&fs.Discovery{},     // directory to search entities
				&github.Discovery{}, // github repo to search entities
			},
			EntityRepo: &store.Repo{}, // set repository plugin configurable by looking at config file. supports pg and sqlite
			//EntityRepo: &pg.Repo{},     //use postgres implementation explicitly. import "github.com/iamgoroot/backline/pkg/store/repo/pg"
			//EntityRepo: &sqlite.Repo{}, //use postgres sqlite explicitly. import "github.com/iamgoroot/backline/pkg/store/repo/sqlite"

			KeyValStore: &store.KV{}, // set key-value storage plugin configurable with config file. supports pg and sqlite
			//KeyValStore: &kv.PgKV{},   // use postgres KV store explicitly. import  "github.com/iamgoroot/backline/pkg/store/kv"
			//KeyValStore: &kv.SqliteKV{}, // use postgres KV store explicitly. import  "github.com/iamgoroot/backline/pkg/store/kv"

			// job scheduler plugin.  Basic implementation that uses KV store and Locker for scheduling and synchronizing tasks
			JobScheduler: &store.Scheduler{},
			// distributed lock plugin configurable with config file.
			// Uses pg_try_advisory_xact_lock for pg or sql table with transaction for sqlite
			DistributedLocker: &store.Locker{},
			// add search plugin based on elastic search. import "github.com/iamgoroot/backline/plugin/search/elastic"
			SearchPlugin: &elastic.Search{},

			// add search plugin based on bluge library (storing indices in local file). import "github.com/iamgoroot/backline/plugin/search/bluge"
			//SearchPlugin: &bluge.Search{},

			ScannerPlugin: &scanner.Plugin{}, // add scanner plugin to scan/read entities.
		},
		Plugins: []core.Plugin{
			catalog.Plugin{}, // add web interface for catalog
			oauth2.Plugin{},  // add oauth2 plugin for catalog authentication
			stock.Theme{},    // add stock theme for catalog

			&techdocs.Plugin{}, // add techdocs documentation plugin

			swaggerui.Plugin{},     // add ability to display openapi specs on entity pages
			asyncapi.Plugin{},      // add ability to display asyncapi specs on entity pages
			rawdefinition.Plugin{}, // add ability to display raw API definitions on entity pages

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
