package plugin

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

type plugins struct {
	htmlHeaders      components
	headerPlugins    components
	sideBarPlugin    components
	entityProcessors []core.EntityProcessor
	entityTab        []core.EntityTabPlugin
	discoveries      []core.Discovery
}

func (p *plugins) Setup(_ context.Context, _ core.Dependencies) error {
	return nil
}

func (p *plugins) HTMLHeader() core.Component {
	return p.htmlHeaders
}

func (p *plugins) HeaderItem() core.Component {
	return p.headerPlugins
}
func (p *plugins) SideBarItem() core.Component {
	return p.sideBarPlugin
}

func (p *plugins) GetEntityTabPlugins() []core.EntityTabPlugin {
	return p.entityTab
}

func (p *plugins) ProcessEntity(ctx context.Context, deps core.Dependencies, entity *model.Entity) error {
	for _, p := range p.entityProcessors {
		if err := p.ProcessEntity(ctx, deps, entity); err != nil {
			return err
		}
	}

	return nil
}
