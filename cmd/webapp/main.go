package main

import (
	"github.com/iamgoroot/backline/plugin/auth/oauth2"
	"github.com/iamgoroot/backline/plugin/documentation/asyncapi"
	"github.com/iamgoroot/backline/plugin/documentation/swaggerui"
	"github.com/iamgoroot/backline/plugin/search/bluge"
	"log/slog"
	"os"

	"github.com/iamgoroot/backline/app"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/store"
	"github.com/iamgoroot/backline/plugin/catalog"
	"github.com/iamgoroot/backline/plugin/documentation/rawdefinition"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs"
	"github.com/iamgoroot/backline/plugin/theme/stock"
)

func main() {
	// start catalog only
	application := app.App{
		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{},
			EntityRepo:        &store.Repo{}, // set repository plugin configurable with config file. supports pg and sqlite
			KeyValStore:       &store.KV{},   // set key-value storage plugin configurable with config file. supports pg and sqlite
			JobScheduler:      &store.Scheduler{},
			DistributedLocker: &store.Locker{},
			SearchPlugin:      &bluge.Search{}, // add search plugin (bluge uses local files for saving indices)
		},
		Plugins: []core.Plugin{
			catalog.Plugin{}, // add web interface for catalog
			oauth2.Plugin{},  // add oauth2 plugin for catalog authentication
			stock.Theme{},    // add stock theme for catalog

			&techdocs.Plugin{}, // add techdocs documentation plugin

			swaggerui.Plugin{},     // add ability to display openapi specs on entity pages
			asyncapi.Plugin{},      // add ability to display asyncapi specs on entity pages
			rawdefinition.Plugin{}, // add ability to display raw API definitions on entity pages
		},
	}

	err := application.Run()

	if err != nil {
		slog.Error("Error while running. Shutting down", slog.Any("err", err))
		os.Exit(1)
	}
}
