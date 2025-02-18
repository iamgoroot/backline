package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/iamgoroot/backline/app/internal/plugin"
	"github.com/iamgoroot/backline/pkg/config"
	"github.com/iamgoroot/backline/pkg/core"
)

func (a *App) initCore(ctx context.Context, cfg config.CfgReader) (*CoreCfg, core.Dependencies, error) {
	coreCfg := getDefaultConfig()

	err := cfg.ReadAt("$.core", coreCfg)
	if err != nil {
		slog.Error("failed to read core config", slog.Any("err", err))
		return nil, nil, err
	}
	// create logger
	logLevel := getLogLvl(coreCfg.Logger)
	logger := createLogger(logLevel, coreCfg.Logger)
	slog.SetDefault(logger)

	// create server
	server := a.createServer(logger, coreCfg)
	deps, err := plugin.InitPlugins(ctx, logger, server, cfg, &a.PluggableDeps, a.Plugins)

	return coreCfg, deps, err
}

func createLogger(level slog.Level, cfg LoggerCfg) *slog.Logger {
	if cfg.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func getLogLvl(cfg LoggerCfg) slog.Level {
	defaultLevel := slog.LevelInfo

	logLevel := &defaultLevel
	if cfg.Level != "" {
		err := logLevel.UnmarshalText([]byte(cfg.Level))
		if err != nil {
			slog.Error("failed to parse log level use one of: ", slog.Any("err", err))
		}
	}

	level := *logLevel

	return level
}
