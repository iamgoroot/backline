package scanner

import (
	"context"
	"log/slog"
	"time"

	"github.com/iamgoroot/backline/plugin/scanner/internal/views"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate internal/views/...

const scannerCfgKey = "$.core.scanner"

type Cfg struct {
	ScanPeriod         time.Duration `yaml:"scanPeriod"`
	EnablePeriodicScan bool          `yaml:"enablePeriodicScan"`
	EnableScanEndpoint bool          `yaml:"enableScanEndpoint"`
	EnableScanButton   bool          `yaml:"enableScanButton"`
	ScanBeforeStart    bool          `yaml:"scanBeforeStart"`
}

type Plugin struct {
	cfg Cfg
}

type EntityProcessor interface {
	ProcessEntity(context.Context, core.Dependencies, *model.Entity) error
}

func (p *Plugin) Setup(ctx context.Context, deps core.Dependencies) error {
	if err := deps.CfgReader().ReadAt(scannerCfgKey, &p.cfg); err != nil {
		return err
	}

	if p.cfg.ScanBeforeStart {
		_, err := p.Scan(ctx, deps)
		if err != nil {
			deps.Logger().Error("failed to scan", slog.Any("err", err))
		}
	}

	if p.cfg.EnablePeriodicScan {
		go func() {
			err := p.ScanPeriodically(ctx, deps, p.cfg.ScanPeriod)
			deps.Logger().Error("failed to start periodic scan", slog.Any("err", err))
		}()
	}

	if p.cfg.EnableScanEndpoint {
		deps.Router().POST("/scanner/run", func(c echo.Context) error {
			ctx := c.Request().Context()
			_, err := p.Scan(ctx, deps) //TODO: result of scan

			return err
		})
	}

	return nil
}

func (p *Plugin) HeaderItem() core.Component {
	showBtn := func() bool {
		return p.cfg.EnableScanButton
	}

	return core.WeighedComponent{Component: views.ScanButton(showBtn), DisplayWeight: 0}
}

func (p *Plugin) Shutdown(_ context.Context) error {
	return nil
}
