package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamgoroot/backline/app/internal/plugin"
	"github.com/iamgoroot/backline/pkg/config"
	"github.com/iamgoroot/backline/pkg/core"
)

var errMissingCertFile = fmt.Errorf("certFile and keyFile must be provided to start https server")

const maxShutdownWait = 30 * time.Second // TODO config

type PluggableDeps = plugin.PluggableDeps

type App struct {
	PluggableDeps
	Plugins []core.Plugin
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	configLocation := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	// read config file
	cfg, err := config.ReadYamlCfgFile(*configLocation)
	if err != nil {
		slog.Error("failed to read config", slog.Any("err", err))
		return err
	}

	// initialize core dependencies
	coreCfg, depsHolder, err := a.initCore(ctx, cfg)
	if err != nil {
		slog.Error("failed to initialize dependencies", slog.Any("err", err))
		return err
	}

	defer func() { // shutdown plugins on exit
		if err = depsHolder.Plugins().Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown plugins", slog.Any("err", err))
		}
	}()

	logger := depsHolder.Logger()

	if coreCfg.Server.Disabled {
		return nil
	}

	go func() {
		<-ctx.Done()
		logger.Info("received shutdown signal or encountered critical error. Shutting down.")

		ctx, done := context.WithTimeout(context.Background(), maxShutdownWait)

		err := depsHolder.Router().Shutdown(ctx)
		if err != nil {
			logger.Error("failed to properly shutdown http server", slog.Any("errMissingCertFile", err))
		}

		done()
	}()

	address := fmt.Sprintf("%s:%d", coreCfg.Server.Address, coreCfg.Server.Port)
	// serve until shutdown
	if coreCfg.Server.HTTPS.Disabled {
		return depsHolder.Router().Start(address)
	}

	httpsCfg := coreCfg.Server.HTTPS

	if httpsCfg.CertFile == "" || httpsCfg.KeyFile == "" {
		return errMissingCertFile
	}

	return depsHolder.Router().StartTLS(address, httpsCfg.CertFile, httpsCfg.KeyFile)
}
