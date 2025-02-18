## Core plugins

### Repository Configuration

The current repository plugin uses Bun to connect to the database. It supports PostgreSQL and SQLite. There are a few ways to initialize the repository plugin:

- `github.com/iamgoroot/backline/pkg/store.Repo` - PostgreSQL or SQLite based on YAML config. Just populate the config with any of the following configurations, and it will create the repository plugin automatically.
- `github.com/iamgoroot/backline/pkg/store/repo/pg.PgRepo` - create a repo specific to PostgreSQL. There are several config options:
  - Pass the config directly into the Config field of this struct:
    ```golang
    repo := pg.PgRepo{
        Config: &pg.Cfg{
            DSN: "postgres://user:password@localhost:5432/backline?sslmode=disable", // use DSN or other options
        }
    }
    ```
  - Use YAML config file:
    ```yaml
    core:
        repo:
            pg:
                dsn: "postgres://user:password@localhost:5432/backline?sslmode=disable"
    ```
- `github.com/iamgoroot/backline/pkg/store/repo/sqlite.SqliteRepo` - create a repo specific to SQLite. Also configurable with YAML or by passing config:
  ```golang
  repo := sqlite.SqliteRepo{
      Config: &sqlite.Cfg{
          DSN: "file:/tmp/backline.db",
      }
  }
  ```
  or
  ```yaml
  core:
      repo:
          sqlite:
              dsn: "file:/tmp/backline.db"
  ```

### Key-Value Storage Configuration

Same approach as repository configuration. Just replace `core.repo` with `core.kv`:

```yaml
core:
  kv:
    sqlite:
      dsn: "file:/tmp/backline.db"
```

### Distributed Lock Configuration

Sometimes plugins need to coordinate between multiple instances, for example, to exclude simultaneous scans. There are a few ways to configure the distributed lock plugin:

- `github.com/iamgoroot/backline/pkg/store/locker.NoOP` - no locking (default)
- `github.com/iamgoroot/backline/pkg/store/locker.PgLock` - use PostgreSQL as distributed lock (uses `pg_try_advisory_xact_lock`)
- `github.com/iamgoroot/backline/pkg/store/locker.SQLLock` - use SQL transaction to lock (intended to be used with SQLite)

### Job Scheduler Configuration

The current job scheduler is minimal. It uses the DistributedLock plugin and key-value store to coordinate between multiple instances.

## Other Plugins

- [catalog](./plugins/catalog) - Main UI for browsing and displaying entities
- [scanner](./plugins/scanner) - Plugin for scanning entities
- [discovery](./plugins/discovery) - Plugins to read entities from different sources (file system, github, etc.)
- [oauth2](./plugins/auth/oauth2) - Plugin for authentication using OAuth2
- [openapiexplorer](./plugins/documentation/openapiexplorer) - Plugin for displaying OpenAPI specs
- [rawdefinition](./plugins/documentation/rawdefinition) - Plugin for displaying raw text API definitions
- [techdocs](./plugins/documentation/techdocs) - Plugin for displaying techdocs
- [default theme](./plugins/theme/stock) - Simple default theme. Supports light and dark mode

## Plugin Development

Backline's plugin system allows you to add custom functionality. Currently, you can add:
- Endpoints for backend in the Setup function
- Header menu items
- Entity tabs
- Sidebar items
