package main

import (
	"github.com/iamgoroot/backline/app"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store"
	"github.com/iamgoroot/backline/pkg/store/kv"
	"github.com/iamgoroot/backline/pkg/store/repo/pg"
	"github.com/iamgoroot/backline/plugin/catalog"
	"github.com/iamgoroot/backline/plugin/discovery/fs"
	"github.com/iamgoroot/backline/plugin/scanner"
	"github.com/iamgoroot/backline/plugin/theme/stock"
	"log"
)

func main() {
	application := app.App{
		// run catalog and scanner

		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{ // Add Location Readers so backline knows how to read entities from different sources
				&fs.Discovery{}, // directory to search entities
			},
			EntityRepo:        &pg.Repo{},         //use postgres implementation explicitly. import "github.com/iamgoroot/backline/pkg/store/repo/pg"
			KeyValStore:       &kv.PgKV{},         // use postgres KV store explicitly. import  "github.com/iamgoroot/backline/pkg/store/kv"
			JobScheduler:      &store.Scheduler{}, // job scheduler plugin. Basic implementation that uses KV store and Locker for scheduling and synchronizing tasks
			DistributedLocker: &store.Locker{},    // distributed lock plugin configurable with config file. Uses pg_try_advisory_xact_lock for pg and sql table with transaction is used for sqlite
			ScannerPlugin:     &scanner.Plugin{},  // add scanner plugin to scan/read entities.
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
