package core

import (
	"context"

	"github.com/iamgoroot/backline/pkg/model"
)

// Plugin is the basic interface that all plugins must implement.
// It provides a way to set up and shutdown the plugin.
type Plugin interface {
	// Setup is called when the plugin is being initialized.
	Setup(rootCtx context.Context, deps Dependencies) error
	// Shutdown is called when the app is being shutdown. It should clean up any resources.
	Shutdown(ctx context.Context) error
}

type Plugins interface {
	HeaderPlugin
	SideBarPlugin
	HTMLHeaderPlugin
	GetEntityTabPlugins() []EntityTabPlugin
	EntityProcessor
}

// HeaderPlugin is a plugin that provides a component to be rendered in the header of the app.
type HeaderPlugin interface {
	Plugin
	// HeaderItem returns a component that will be rendered in the header of the app.
	HeaderItem() Component
}

// SideBarPlugin is a plugin that provides a component to be rendered in the sidebar of the app.
type SideBarPlugin interface {
	Plugin
	// SideBarItem returns a component that will be rendered in the sidebar of the app.
	SideBarItem() Component
}

// HTMLHeaderPlugin is a plugin that provides a component to be rendered in the HTML header of the app.
type HTMLHeaderPlugin interface {
	Plugin
	// HTMLHeader returns a component that will be rendered in the HTML header of the app.
	HTMLHeader() Component
}

// EntityTabPlugin is a plugin that provides a component to be rendered as a tab on the entity page.
type EntityTabPlugin interface {
	Plugin
	// EntityTabLink returns a component that will be rendered as a tab on the entity page.
	// Linked page should be rendered by the plugin itself using Router() from Dependencies.
	EntityTabLink(entity *model.Entity) Component
}

// EntityProcessor is a plugin that processes entities.
type EntityProcessor interface {
	Plugin
	// ProcessEntity is called when a new entity is discovered or updated.
	ProcessEntity(ctx context.Context, deps Dependencies, entity *model.Entity) error
}

// NoOpShutdown is way to mark a plugin that does not need any cleanup or shutdown routine.
type NoOpShutdown struct{}

func (NoOpShutdown) Shutdown(_ context.Context) error { return nil }
