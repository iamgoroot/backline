package scanner

import (
	"context"
	"errors"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
	"golang.org/x/sync/semaphore"
)

func (p *Plugin) ScanPeriodically(ctx context.Context, deps core.Dependencies, scanPeriod time.Duration) error {
	ticker := time.NewTicker(scanPeriod)
	defer ticker.Stop()

	return deps.Scheduler().WithTimeout(
		ctx, "backline-scan", scanPeriod, func(ctx context.Context) error {
			deps.Logger().Info("starting scan", slog.Any("period", scanPeriod.String()))
			_, err := p.Scan(ctx, deps)

			return err
		},
	)
}

func (p *Plugin) Scan(ctx context.Context, deps core.Dependencies) (ScanResults, error) {
	var results ScanResults
	err := deps.DistributedLock().WithLock(ctx, "backline-scan-lock", func(ctx context.Context) error {
		var err error
		results, err = p.scan(ctx, deps)
		return err
	})
	return results, err
}

func (p *Plugin) scan(ctx context.Context, deps core.Dependencies) (ScanResults, error) {
	registrator := entityRegistrator{ctx: ctx, deps: deps, registeredNames: map[string]struct{}{}}

	discoveryGroup := sync.WaitGroup{}
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, discoverer := range deps.Discoveries() {
		discoveryGroup.Add(1)

		err := sem.Acquire(ctx, 1)

		if err != nil {
			deps.Logger().Error("failed to acquire semaphore", slog.Any("err", err))
			return ScanResults{}, err
		}

		go func() {
			defer sem.Release(1)
			defer discoveryGroup.Done()

			err := discoverer.Discover(ctx, deps, registrator.RegisterEntity)
			if err != nil {
				deps.Logger().Error("failed to process location", slog.Any("err", err))
			}
		}()
	}

	discoveryGroup.Wait()

	results, err := p.collectResults(ctx, deps, registrator)
	if err != nil {
		return results, err
	}
	return p.updateEntities(ctx, deps, results)
}

// updateEntities updates entity relations, orphan status, removes indices. TODO: make it event driven
func (p *Plugin) updateEntities(ctx context.Context, deps core.Dependencies, results ScanResults) (ScanResults, error) {
	if len(results.FoundEntities) > 0 {
		err := deps.Repo().UpdateRelations(ctx, results.FoundEntities...)
		if err != nil {
			return results, err
		}
	}

	if len(results.OrphanedEntities) > 0 {
		deps.Logger().Info("found orphaned entities", slog.Any("entities", results.OrphanedEntities))
		now := time.Now()
		err := errors.Join(
			deps.Repo().UpdateAsOrphans(ctx, &now, results.OrphanedEntities...),
			deps.Search().RemoveIndex(ctx, results.OrphanedEntities...),
		)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

func (p *Plugin) collectResults(ctx context.Context, deps core.Dependencies, registrator entityRegistrator) (ScanResults, error) {
	foundEntities := make([]string, 0, len(registrator.registeredNames))
	for entityName := range registrator.registeredNames {
		foundEntities = append(foundEntities, entityName)
	}

	previousEntities := make([]string, 0, len(registrator.registeredNames)+10) // assume that there are some deleted entities to allocate mem
	err := deps.StoreKV().Get(ctx, "scanner", "entityList", &previousEntities)
	if err != nil {
		deps.Logger().Info("could not retrieve previous entity list", slog.Any("err", err))
	}
	err = deps.StoreKV().Set(ctx, "scanner", "entityList", &foundEntities, 0)
	if err != nil {
		deps.Logger().Info("could not store entity list", slog.Any("err", err))
	}
	orphanedEntities := make([]string, 0, 10)
	for _, entityName := range previousEntities {
		if _, ok := registrator.registeredNames[entityName]; !ok {
			orphanedEntities = append(orphanedEntities, entityName)
		}
	}

	return ScanResults{
		FoundEntities:    foundEntities,
		OrphanedEntities: orphanedEntities,
	}, nil
}

type ScanResults struct {
	FoundEntities    []string
	OrphanedEntities []string
}
