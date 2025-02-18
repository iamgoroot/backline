package locker

import (
	"context"

	"github.com/iamgoroot/backline/pkg/core"
)

type NoOP struct {
	core.NoOpShutdown
}

func (m NoOP) Setup(_ context.Context, _ core.Dependencies) error {
	return nil
}
func (m NoOP) WithLock(ctx context.Context, _ string, run func(ctx context.Context) error) error {
	return run(ctx)
}
