package plugin

import (
	"context"
	"io"

	"github.com/iamgoroot/backline/pkg/core"
)

type components []core.Component

func (c components) Render(ctx context.Context, w io.Writer) error {
	for _, comp := range c {
		if err := comp.Render(ctx, w); err != nil {
			return err
		}
	}

	return nil
}

func (c components) Weight() int {
	return 0
}
