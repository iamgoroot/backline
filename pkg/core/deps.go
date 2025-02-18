package core

import (
	"context"
	"time"

	"github.com/iamgoroot/backline/pkg/model"
	"github.com/labstack/echo/v4"
)

type ctxKey string

const (
	CSRFTokenContextKey ctxKey = "csrfTokenCtxKey" // nolint gosec
	CSRFQueryParamName  string = "csrf"
)

func GetCSRFToken(ctx context.Context) string {
	value := ctx.Value(CSRFTokenContextKey)
	if value == nil {
		return ""
	}

	csrf, _ := value.(string)

	return csrf
}

type Dependencies interface {
	// Router returns the echo router to be used by the plugin to register http routes.
	Router() *echo.Echo // unfortunately cannot be abstract. TODO: decide later whether to use default golang http server
	// Logger returns the logger to be used by the plugin to log messages.
	Logger() Logger
	// Repo returns the repository to be used by the plugin to store and read entities.
	Repo() Repo
	// CfgReader returns the configuration reader to be used by the plugin to read configuration values.
	CfgReader() CfgReader
	// DistributedLock returns the distributed lock manager to be used by the plugin to coordinate between multiple instances.
	DistributedLock() DistributedLock
	// Scheduler returns the job scheduler to be used by the plugin to run tasks with specific intervals or at specific times.
	Scheduler() Scheduler
	// Discoveries returns the list of discovery plugins to be used by the plugin to read entities from different sources.
	Discoveries() []Discovery
	// Plugins returns the configured UI plugins to be used by the plugin to extend the UI with new functionality.
	Plugins() Plugins
	// StoreKV returns the key-value store to be used by the plugin to store and read data that can be used by other plugins.
	StoreKV() StoreKV
	// Search returns the search plugin to be used by the plugin to search and index entities.
	Search() Search
}

type CfgReader interface {
	// ReadAt reads the configuration values from the specified path and stores them in the provided configuration object.
	// It returns an error if the configuration values cannot be read successfully.
	// Example:
	// ```
	// type Config struct {
	//	Val string `yaml:"val"`
	// }
	// var cfg Config
	// err := reader.ReadAt("$.key.subkey", &cfg)
	// ```
	// This will allow to read following yaml config:
	// ```
	// key:
	//   subkey:
	//     val: env:ENV_VAR_NAME|dummy_value
	// ```
	// In the example above, `Val` field will be set to the value of the environment variable `ENV_VAR_NAME` or `"dummy_value"` if env not set.
	ReadAt(path string, cfg any) error
}
type Logger interface {
	Error(string, ...any)
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}
type Repo interface {
	Plugin
	Store(ctx context.Context, entity *model.Entity) error
	List(ctx context.Context, params *model.ListEntityReq) ([]*model.Entity, error)
	GetByName(ctx context.Context, name string) (*model.Entity, error)
	UpdateRelations(ctx context.Context, entityNames ...string) error
	UpdateAsOrphans(ctx context.Context, at *time.Time, entityNames ...string) error
}

type DistributedLock interface {
	Plugin
	// WithLock runs the provided function with the specified lock key and returns an error if functions is being executed by another instance.
	WithLock(ctx context.Context, lockKey string, run func(ctx context.Context) error) error
}
type Scheduler interface {
	Plugin
	// WithTimeout runs the provided function with the specified interval.
	// Does not run on this instance if function with given lockKey is already running on another instance (if proper locker is used)
	WithTimeout(ctx context.Context, lockKey string, timeout time.Duration, run func(ctx context.Context) error) error
}

type RegistrationFunc func(locationMeta *model.LocationMetadata, entity *model.Entity)
type Discovery interface {
	Plugin
	// Discover reads entities from the specified location and registers them using the provided registration function.
	Discover(ctx context.Context, deps Dependencies, register RegistrationFunc) error
	// TryDownload tries to download the entity from the specified location and returns body of downloaded entity.
	// Returns an error if the entity cannot be downloaded from given location.
	TryDownload(ctx context.Context, deps Dependencies, entity *model.LocationMetadata, ref string) (string, error)
}

type Scanner interface {
	Plugin
}

type StoreKV interface {
	Plugin
	// Set stores the provided value with the specified key in the specified key group with the specified TTL.
	Set(ctx context.Context, keyGroup, key string, value any, ttl time.Duration) error
	// Get retrieves the value with the specified key from the specified key group and populates the provided value object.
	Get(ctx context.Context, keyGroup, key string, value any) error
}

type Search interface {
	Plugin
	Search(ctx context.Context, query string, offset, limit int) ([]SearchResult, error)
	Index(ctx context.Context, entityName string, reference string, category string, value string) error
	RemoveIndex(ctx context.Context, entityName ...string) error
}
type SearchResult struct {
	EntityName string
	Link       string
	Category   string
	Highlight  string
}
