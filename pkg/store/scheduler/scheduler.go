package scheduler

import (
	"context"
	"errors"
	"time"

	"github.com/iamgoroot/backline/pkg/core"
)

type Scheduler struct {
	core.NoOpShutdown
	deps core.Dependencies
}

func (m *Scheduler) Setup(_ context.Context, deps core.Dependencies) error {
	m.deps = deps
	return nil
}

func (m *Scheduler) WithTimeout(ctx context.Context, jobName string, timeout time.Duration, run func(ctx context.Context) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		jobFunc := m.createJobFunc(jobName, timeout, run)
		err := m.deps.DistributedLock().WithLock(ctx, jobName, jobFunc)

		var lockTaken core.ErrLockTaken

		switch {
		case errors.Is(err, lockTaken):
			m.deps.Logger().Error("skipping job. lock is already taken by another process", "jobName", jobName, "error", err)
		case err != nil:
			m.deps.Logger().Error("failed to run job", "jobName", jobName, "error", err)
			return err
		default:
			m.deps.Logger().Info("job completed", "jobName", jobName)
		}
	}
}

func (m *Scheduler) createJobFunc(
	jobName string, timeout time.Duration, run func(ctx context.Context) error,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var lastRun time.Time

		err := m.deps.StoreKV().Get(ctx, "job", jobName, &lastRun)
		if err != nil {
			var NotFoundError core.NotFoundError
			if !errors.Is(err, NotFoundError) {
				return err
			}
		}

		if !lastRun.IsZero() {
			lastInterval := time.Since(lastRun)
			if lastInterval < timeout {
				sleepFor := timeout - lastInterval
				sleepCtx(ctx, sleepFor) // sleep until timeout is reached

				return nil
			}
		}

		err = run(ctx)
		if err != nil {
			return err
		}

		currentTime := time.Now()

		return m.deps.StoreKV().Set(ctx, "job", jobName, &currentTime, 0)
	}
}

func sleepCtx(ctx context.Context, duration time.Duration) {
	for duration > 0 {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			duration -= time.Second
		}
	}
}
