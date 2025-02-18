package internal

import (
	"context"
	"errors"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

func NewTaskGroup(ctx context.Context) (*TaskGroup, context.Context) {
	errGroup, groupCtx := errgroup.WithContext(ctx)
	errGroup.SetLimit(runtime.NumCPU())

	return &TaskGroup{Group: errGroup, m: sync.Mutex{}}, groupCtx
}

// TaskGroup is a wrapper around errgroup.Group with error aggregation.
type TaskGroup struct {
	*errgroup.Group
	AggregatedErr error
	m             sync.Mutex
}

func (g *TaskGroup) Go(f func() error) {
	g.Group.Go(func() error {
		err := f()
		if err != nil {
			g.m.Lock()
			defer g.m.Unlock()
			g.AggregatedErr = errors.Join(g.AggregatedErr, err)
		}

		return err
	})
}

func (g *TaskGroup) Wait() error {
	_ = g.Group.Wait()
	return g.AggregatedErr
}
