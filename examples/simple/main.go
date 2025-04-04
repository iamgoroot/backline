package main

import (
	"log"

	"github.com/iamgoroot/backline/app"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store"
	"github.com/iamgoroot/backline/pkg/store/kv"
	"github.com/iamgoroot/backline/pkg/store/repo/pg"
	"github.com/iamgoroot/backline/plugin/catalog"
	"github.com/iamgoroot/backline/plugin/discovery/fs"
	"github.com/iamgoroot/backline/plugin/scanner"
	"github.com/iamgoroot/backline/plugin/theme/stock"
)

func main() {
	application := app.App{
		// run catalog and scanner

		PluggableDeps: app.PluggableDeps{
			// add Location Readers so backline knows how to read entities from different sources
			EntityDiscoveries: []core.Discovery{
				&fs.Discovery{}, // directory to search entities
			},
			EntityRepo: &pg.Repo{}, // use postgres implementation explicitly.
			// use postgres KV store explicitly.
			KeyValStore: &kv.PgKV{},
			// job scheduler plugin. Basic implementation that uses KV store and Locker for scheduling and synchronizing tasks
			JobScheduler: &store.Scheduler{},
			// distributed lock plugin configurable with config file.
			// uses pg_try_advisory_xact_lock for pg or sql table with transaction for sqlite
			DistributedLocker: &store.Locker{},
			// add scanner plugin to scan/read entities.
			ScannerPlugin: &scanner.Plugin{},
		},
		Plugins: []core.Plugin{
			catalog.Plugin{}, // add web interface for catalog
			stock.Theme{},    // add stock theme for catalog
		},
	}

	err := application.Run()
	if err != nil {
		log.Fatal(err)
	}
}
