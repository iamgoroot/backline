package oauth2

import (
	"context"
	"io"
	"math"

	"github.com/iamgoroot/backline/pkg/core"
)

var headerItems = core.ComponentFunc(func(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, logoutHTML)
	return err
}, math.MaxInt)

func (Plugin) HeaderItem() core.Component {
	return headerItems
}
