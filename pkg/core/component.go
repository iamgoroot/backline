package core

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
)

type Component interface {
	Render(ctx context.Context, w io.Writer) error
	Weight() int
}

type StyledComponent struct {
	Component
	Class string
}

func (c StyledComponent) Render(ctx context.Context, writer io.Writer) error {
	if _, err := fmt.Fprintf(writer, `<div class="%s">`, c.Class); err != nil {
		return err
	}

	if err := c.Component.Render(ctx, writer); err != nil {
		return err
	}

	_, err := fmt.Fprint(writer, `</div>`)

	return err
}

func ComponentFunc(f func(ctx context.Context, w io.Writer) error, weight int) Component {
	return componentFunc{Func: f, DisplayWeight: weight}
}

type componentFunc struct {
	Func          func(ctx context.Context, w io.Writer) error
	DisplayWeight int
}

func (c componentFunc) Render(ctx context.Context, w io.Writer) error {
	return c.Func(ctx, w)
}

func (c componentFunc) Weight() int {
	return c.DisplayWeight
}

type WeighedComponent struct {
	templ.Component
	DisplayWeight int
}

func (c WeighedComponent) Render(ctx context.Context, w io.Writer) error {
	return c.Component.Render(ctx, w)
}

func (c WeighedComponent) Weight() int {
	return c.DisplayWeight
}

type EmptyComponent struct {
}

func (c EmptyComponent) Render(_ context.Context, _ io.Writer) error {
	return nil
}

func (c EmptyComponent) Weight() int {
	return 0
}

func ErrComponent(err error) Component {
	return errComponent{Error: err}
}

type errComponent struct {
	Error error
}

func (err errComponent) Render(_ context.Context, _ io.Writer) error {
	return err.Error
}
func (err errComponent) Weight() int {
	return 0
}
