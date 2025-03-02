# Backline IDP

Backline is a minimalistic Internal Developer Portal (IDP) inspired by Backstage, built using Go and HTMX. It provides a flexible and extensible platform for exploring various developer resources.

## Table of Contents
- [Features](#features)
- [Why?](#why)
- [Getting Started](#getting-started)
- [Customization](#customization)
- [Configuration](#configuration)
  - [Basics](#basics)
  - [Core Config](#core-config)
  - [Repository Configuration](#repository-configuration)
  - [Key-Value Storage Configuration](#key-value-storage-configuration)
  - [Distributed Lock Configuration](#distributed-lock-configuration)
  - [Job Scheduler Configuration](#job-scheduler-configuration)
- [Existing Plugins](#existing-plugins)
- [Plugin Development](#plugin-development)
- [Contributing](#contributing)
- [License](#license)

# Demo

Check out deployed demo app loaded with same entities that backstage use for their demo app and backstage itself:

[https://backline-demo.koyeb.app](https://backline-demo.koyeb.app)

[https://backline.onrender.com](https://backline.onrender.com)

Some features may be disabled (for example "Scan" button)

## Features
- **Use existing backstage entities**: Backline can be used with your existing Backstage entity definitions, so you can use your existing entity definitions and metadata.
- **No scaffolding**: You don't need to worry about setting up a complex backend or frontend, Backline handles it all for you. (see "./cmd" directory for examples)
- **Plugin System**: Backline supports a simple, type-safe plugin system, allowing developers to extend and customize the portal's functionality with ease. (see "./plugins" directory)
- **OAuth2 Integration**: Authentication and authorization using OAuth2.

## Why?
- **Lower learning curve**: Backline is designed to be easier to learn and use.
- **Simpler upgrades**: Unlike Backstage, which requires scaffolding and is hard to upgrade, Backline isn't scaffolded and should work just by updating go.mod.
- **Easy plugin development**:
  - Plugins register themselves and do not require any code modifications.
  - Plugins implement statically typed interfaces, making them easy to implement and detect version incompatibilities.
- **Addressing Backstage issues**:
  - Backstage does not support search on openapi and asyncapi specs (closed as not planned [backstage/backstage#22802](https://github.com/backstage/backstage/issues/22802)).
  - Backstage does not respect some relations like `apiProvidedBy` out of the box [backstage/backstage#25387](https://github.com/backstage/backstage/issues/25387).
  - Plugins use different HTTP clients that may not respect proxy settings [help-im-behind-a-corporate-proxy](https://github.com/backstage/backstage/blob/master/contrib/docs/tutorials/help-im-behind-a-corporate-proxy.md).

## Getting Started

To get started with Backline you need to add as least few plugins or specify implementations for core functionality

Lets start with a simple example of running just catalog and scanner with minimalistic set of plugins

We'll use:

* Postgres for storage (both entity and as Key-Value) and distributed lock (using pg_try_advisory_xact_lock)


```golang
	application := app.App{
		// run catalog and scanner

		PluggableDeps: app.PluggableDeps{
			EntityDiscoveries: []core.Discovery{    // Add Location Readers so backline knows how to read entities from different sources
				&fs.Discovery{},                      // github repo to search entities
			},
			EntityRepo:         &pg.Repo{},         //use postgres implementation explicitly.
			KeyValStore:        &kv.PgKV{},         // use postgres KV store explicitly.
			JobScheduler:       &store.Scheduler{}, // job scheduler plugin. Basic implementation that uses KV store and Locker for scheduling and synchronizing tasks
			DistributedLocker:  &store.Locker{},    // distributed lock plugin configurable with config file. Uses pg_try_advisory_xact_lock for pg and sql table with transaction is used for sqlite
			ScannerPlugin:      &scanner.Plugin{},  // add scanner plugin to scan/read entities.
		},
		Plugins: []core.Plugin{
			catalog.Plugin{},                       // add web interface for catalog
			stock.Theme{},                          // add stock theme for catalog
		},
	}

  err := application.Run()
  if err != nil {
    log.Fatal(err)
  }
```

Although configuration can be done by populating plugin fields directly in the code, we will use a configuration file
For sake of simplicity we will disable some security features (https, csrf, cors are enabled by default)


```yaml
core:
  server:
    https:
      disabled: true
    csrf:
      disabled: true
    cors:
      disabled: true
  logger:
    level: env:LOG_LEVEL
    format: json
  repo:
    pg:
      dsn: env:PG_DSN
  kv:
    pg:
      dsn: env:PG_DSN
  lock:
    pg:
      dsn: env:PG_DSN
  scanner:
    enableScanEndpoint: true
    enableScanButton: true
locations:
  fs:
    - "./entities"

```

You can populate config file directly or use env variables for config values

```env
PG_DSN=postgresql://postgres:postgres@localhost:5432/backline?sslmode=disable
LOG_LEVEL=INFO
```

Run the application

```bash
go run main.go --config {your-config-location}/config.yaml
```

Open [http://localhost:8080](http://localhost:8080) and you should see the catalog UI. Click on `Scan entities` button to start scanning entities in directory `./entities`

See more examples in `./examples` directory


## Customization

You can customize the appearance and functionality of Backline by modifying the configuration file or creating your own set of plugins.

### Run separate scanner in a separate service

To run entity scan in a separate process:

```bash
go run github.com/iamgoroot/backline/examples/scan_service --config {your-config-location}/config.yaml
```

Or run entity scan as a library:

```golang
application := app.App{
    Plugins: []app.Plugin{
        &scanner.Plugin{},
    },
}
err := application.Run()
```

### Run catalog UI separately

To run catalog UI in a separate process, run:

```bash
go run github.com/iamgoroot/backline/examples/webapp --config {your-config-location}/config.yaml
```

Or run catalog UI as a library in an existing service:

```golang
application := app.App{
    Plugins: []app.Plugin{
        catalog.Plugin{},
        stock.Theme{},
    },
}

err := application.Run()
```

This will start Backline with no plugins except the default UI theme and catalog plugin. No OAuth2, no scanner, no OpenAPI explorer, no discovery.

## Configuration

### Basics

Backline uses a YAML config file to configure itself. The config file allows you to specify default values and environment variable overrides.

```yaml
key: value1
key_from_env: env:ENV_VAR_NAME
key_with_default: env:ENV_VAR_NAME|default_value
```

In the example above, `key` will be set to `value1`, `key_from_env` will be set to the value of the environment variable `ENV_VAR_NAME`, and `key_with_default` will be set to the value of the environment variable `ENV_VAR_NAME` or `default_value` if the environment variable is not set.

### Core Config

Here's an example of a minimalistic core config:

```yaml
core:
  server: # configure server ports, host, origins etc
    port: env:PORT|8080
    host: localhost
    origins:
    - http://localhost:8080
  logger: # configure logging level and format
    level: env:LOG_LEVEL|DEBUG
    format: json
  repo: # configure repository implementation
    sqlite:
      url: ":memory:" # use in-memory SQLite database for the simplest setup
  kv: # configure key-value implementation
    sqlite:
      url: ":memory:" 
```

## Core dependencies and Plugins

For more info about plugins and how to write them, see [plugin readme](./plugin/README.md)

## Contributing

Contributions to Backline are welcome! If you'd like to contribute, please fork the repository and submit a pull request with your changes.

## License

Backline is open-source software licensed under the MIT License.
