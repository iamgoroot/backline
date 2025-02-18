package plugin

import (
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
)

type Dependencies struct {
	*PluggableDeps
	Logging      core.Logger
	ConfigReader core.CfgReader
	EchoRouter   *echo.Echo
}

// Router returns an echo router that can be used by plugins to register new API endpoints.
func (d *Dependencies) Router() *echo.Echo {
	return d.EchoRouter
}

// Logger returns a logger that can be used by plugins to log messages.
func (d *Dependencies) Logger() core.Logger {
	return d.Logging
}

// CfgReader returns a configuration reader that can be used to read configuration values dynamically at Setup time or during runtime.
// Configuration values are read from a config file.
// Config file is a yaml file with ability to specify default values and environment variable overrides
// Example:
// ```
// key: value1
// key_from_env: env:ENV_VAR_NAME
// key_with_default: env:ENV_VAR_NAME|default_value
// ```
// In the example above, `key` will be set to `value1`,
// `key_from_env` will be set to the value of the environment variable `ENV_VAR_NAME`
// `key_with_default` will be set to the value of the environment variable `ENV_VAR_NAME` or `default_value` if env not set.
func (d *Dependencies) CfgReader() core.CfgReader {
	return d.ConfigReader
}

// Plugins returns configured UI plugins that can be used to extend the UI with new functionality.
// Plugins can add new sidebar items, header items, entity page tabs as well as respective API endpoints.
func (d *Dependencies) Plugins() core.Plugins {
	return d.appPlugins
}

type PluggableDeps struct {
	EntityRepo        core.Repo
	SearchPlugin      core.Search
	KeyValStore       core.StoreKV
	DistributedLocker core.DistributedLock
	JobScheduler      core.Scheduler
	ScannerPlugin     core.Scanner
	appPlugins        *plugins
	EntityDiscoveries []core.Discovery
}

// Discoveries returns configured discovery plugins that can read entities from different sources (files, APIs, etc.)
func (d *PluggableDeps) Discoveries() []core.Discovery {
	return d.EntityDiscoveries
}

// StoreKV returns configured key-value store to use by plugins that need to store and read data.
func (d *PluggableDeps) StoreKV() core.StoreKV {
	return d.KeyValStore
}

// Repo returns configured repository for storing and reading entities.
func (d *PluggableDeps) Repo() core.Repo {
	return d.EntityRepo
}

// DistributedLock returns configured distributed lock manager for plugins that need to coordinate between multiple instances.
func (d *PluggableDeps) DistributedLock() core.DistributedLock {
	return d.DistributedLocker
}

// Scheduler returns configured job scheduler for plugins that need to run tasks with specific intervals or at specific times.
func (d *PluggableDeps) Scheduler() core.Scheduler {
	return d.JobScheduler
}

func (d *PluggableDeps) Search() core.Search {
	return d.SearchPlugin
}

func (d *PluggableDeps) Scanner() core.Scanner {
	return d.ScannerPlugin
}
