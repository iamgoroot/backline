package plugin

import (
	"context"
	"errors"
	"slices"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/labstack/echo/v4"
)

func InitPlugins(
	ctx context.Context, logger core.Logger, server *echo.Echo, cfg core.CfgReader, pluggableDeps *PluggableDeps, appPlugins []core.Plugin,
) (*Dependencies, error) {
	plugins := &plugins{}

	processAppPlugins(plugins, appPlugins...) // set all app plugins

	// discover app plugins mixes with core plugins
	processAppPlugins(plugins, pluggableDeps.StoreKV())
	processAppPlugins(plugins, pluggableDeps.Repo())
	processAppPlugins(plugins, pluggableDeps.DistributedLock())
	processAppPlugins(plugins, pluggableDeps.Scheduler())
	processAppPlugins(plugins, pluggableDeps.Search())
	processAppPlugins(plugins, pluggableDeps.Scanner())

	for _, discovery := range pluggableDeps.EntityDiscoveries {
		processAppPlugins(plugins, discovery)
	}

	slices.SortFunc(plugins.sideBarPlugin, compareByWeight) // sort sidebar items by weight
	slices.SortFunc(plugins.headerPlugins, compareByWeight) // sort header items by weight

	pluggableDeps.appPlugins = plugins
	deps := &Dependencies{
		EchoRouter:    server,
		Logging:       logger,
		ConfigReader:  cfg,
		PluggableDeps: pluggableDeps,
	}

	err := errors.Join( // setup plugins
		setupPlugin(ctx, deps, deps.StoreKV()),
		setupPlugin(ctx, deps, deps.Repo()),
		setupPlugin(ctx, deps, deps.DistributedLock()),
		setupPlugin(ctx, deps, deps.Scheduler()),
		setupPlugin(ctx, deps, deps.Search()),
		setupPlugins(ctx, deps, deps.Discoveries()...),
		setupPlugins(ctx, deps, appPlugins...),
	)
	if err != nil {
		return nil, err // do not set up scan if other plugins failed
	}

	err = setupPlugins(ctx, deps, deps.Scanner())

	return deps, err
}

func processAppPlugins(plugins *plugins, appPlugins ...core.Plugin) {
	for _, plug := range appPlugins {
		if plug == nil {
			continue
		}

		if header, ok := plug.(core.HeaderPlugin); ok {
			plugins.headerPlugins = append(plugins.headerPlugins, core.StyledComponent{Component: header.HeaderItem(), Class: "headerItem"})
		}

		if sideBar, ok := plug.(core.SideBarPlugin); ok {
			plugins.sideBarPlugin = append(plugins.sideBarPlugin, sideBar.SideBarItem())
		}

		if htmlHeader, ok := plug.(core.HTMLHeaderPlugin); ok {
			plugins.htmlHeaders = append(plugins.htmlHeaders, htmlHeader.HTMLHeader())
		}

		if entityTab, ok := plug.(core.EntityTabPlugin); ok {
			plugins.entityTab = append(plugins.entityTab, entityTab)
		}

		if entityProcessor, ok := plug.(core.EntityProcessor); ok {
			plugins.entityProcessors = append(plugins.entityProcessors, entityProcessor)
		}

		if discovery, ok := plug.(core.Discovery); ok {
			plugins.discoveries = append(plugins.discoveries, discovery)
		}
	}
}
func setupPlugin(ctx context.Context, deps core.Dependencies, plugin core.Plugin) error {
	if plugin == nil {
		return nil
	}

	return plugin.Setup(ctx, deps)
}
func setupPlugins[Plugin core.Plugin](ctx context.Context, deps core.Dependencies, plugins ...Plugin) error {
	var err error

	for _, plugin := range plugins {
		setupErr := setupPlugin(ctx, deps, plugin)

		err = errors.Join(err, setupErr)
	}

	return err
}

func compareByWeight(a, b core.Component) int {
	weightA := a.Weight()
	weightB := b.Weight()

	if weightA == weightB {
		return 0
	}

	if weightA > weightB {
		return 1
	}

	return -1
}
