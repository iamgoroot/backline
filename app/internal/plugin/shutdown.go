package plugin

import (
	"context"
	"time"

	sync "github.com/iamgoroot/backline/app/internal/sync"
	"github.com/iamgoroot/backline/pkg/core"
)

const maxShutdownWait = 30 * time.Second // TODO config

func (p *plugins) Shutdown(_ context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), maxShutdownWait)
	defer cancel()

	shutdownGroup, shutdownCtx := sync.NewTaskGroup(shutdownCtx)
	shutdownAsync(shutdownCtx, shutdownGroup, p.discoveries...)
	shutdownAsync(shutdownCtx, shutdownGroup, p.entityProcessors...)
	shutdownAsync(shutdownCtx, shutdownGroup, p.entityTab...)

	return shutdownGroup.Wait()
}

func (d *PluggableDeps) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, maxShutdownWait)
	defer cancel()

	shutdownGroup, shutdownCtx := sync.NewTaskGroup(shutdownCtx)
	shutdownAsync(shutdownCtx, shutdownGroup, d.appPlugins)
	shutdownAsync(shutdownCtx, shutdownGroup, d.Discoveries()...)
	shutdownAsync(shutdownCtx, shutdownGroup, d.Scheduler())
	shutdownAsync(shutdownCtx, shutdownGroup, d.DistributedLock())
	shutdownAsync(shutdownCtx, shutdownGroup, d.Repo())
	shutdownAsync(shutdownCtx, shutdownGroup, d.StoreKV())

	return shutdownGroup.Wait()
}

func shutdownAsync[P core.Plugin](ctx context.Context, shutdownGroup *sync.TaskGroup, plugin ...P) {
	for _, shutdowner := range plugin {
		shutdownGroup.Go(func() error {
			return shutdowner.Shutdown(ctx)
		})
	}
}
