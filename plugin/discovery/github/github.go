package github

import (
	"context"
	"crypto/sha1" // nolint gosec not used for security purposes
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var configPath = "$.locations.github"

type Discovery struct {
	core.NoOpShutdown
	cfgs []Config
}

type Config struct {
	Owner       string   `yaml:"owner"`
	Repo        string   `yaml:"repo"`
	Branch      string   `yaml:"branch"`
	AccessToken string   `yaml:"accessToken"`
	Host        string   `yaml:"host"`
	Paths       []string `yaml:"paths"`
	Depth       int      `yaml:"depth"`
}

func (d *Discovery) Setup(_ context.Context, _ core.Dependencies) error { return nil }

func (d *Discovery) Discover(ctx context.Context, deps core.Dependencies, register core.RegistrationFunc) error {
	deps.Logger().Info("starting github discovery")

	err := deps.CfgReader().ReadAt(configPath, &d.cfgs)
	if err != nil {
		return err
	}

	for _, cfg := range d.cfgs {
		paths := cfg.Paths
		if len(paths) == 0 {
			paths = []string{""}
		}

		for _, path := range paths {
			err = d.discoverPath(ctx, deps, &cfg, register, path)
			if err != nil {
				deps.Logger().Error("failed to discover", slog.String("path", path), slog.Any("err", err))
			}
		}
	}

	deps.Logger().Info("finished github discovery")

	return nil
}

func (d *Discovery) discoverPath(
	ctx context.Context, deps core.Dependencies, gitCfg *Config, register core.RegistrationFunc, path string,
) error {
	contentGetter := &githubContentGetter{
		client: getClient(ctx, gitCfg),
		cfg:    gitCfg,
	}
	latestProcessed := d.latestProcessedVersion(ctx, deps, gitCfg, path) //TODO: move up
	latest := contentGetter.getLatestVersion(ctx)

	if latestProcessed == latest {
		return nil
	}

	err := contentGetter.processEntities(ctx, path, 0, register)
	if err != nil {
		return err
	}

	err = deps.StoreKV().Set(ctx, "github-discovery", versionKey(gitCfg, path), latest, 7*24*time.Hour)
	if err != nil {
		deps.Logger().Error("failed to set commit path", slog.Any("err", err))
	}

	return nil
}

func (d *Discovery) latestProcessedVersion(ctx context.Context, deps core.Dependencies, gitCfg *Config, path string) string {
	var val string

	err := deps.StoreKV().Get(ctx, "github-discovery", versionKey(gitCfg, path), &val)
	if err != nil {
		deps.Logger().Debug("failed to get commit path", slog.Any("err", err))
		return ""
	}

	return val
}
func versionKey(gitCfg *Config, path string) string {
	cfgStr, _ := json.Marshal(gitCfg)
	checksum := sha1.Sum(cfgStr) // nolint gosec not used for security purposes

	return fmt.Sprintf("github-discovery-version:%s:path:%s", checksum, path)
}

func getClient(ctx context.Context, cfg *Config) *githubv4.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.AccessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)

	if cfg.Host != "" {
		endpoint := fmt.Sprintf("https://api.%s/graphql", cfg.Host)
		return githubv4.NewEnterpriseClient(endpoint, httpClient)
	}

	return githubv4.NewClient(httpClient)
}
