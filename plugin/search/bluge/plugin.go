package bluge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/plugin/search/ui"
)

const (
	configKey           = "$.core.search.bluge"
	readerReopenTimeout = 5 * time.Minute
)

var _ core.Search = &Search{}

type Search struct {
	cfg bluge.Config
	ui.SearchView
	reader       *bluge.Reader
	writer       *bluge.Writer
	readerTicker *time.Ticker
	Config
	mx sync.Mutex
}

type Config struct {
	Location string `yaml:"location"`
}

func (plugin *Search) Setup(ctx context.Context, deps core.Dependencies) error {
	err := deps.CfgReader().ReadAt(configKey, &plugin.Config)
	if err != nil {
		return err
	}

	if plugin.Location == "" {
		plugin.Location = "./bluge-index"
	}

	plugin.readerTicker = time.NewTicker(readerReopenTimeout)
	plugin.cfg = bluge.DefaultConfig(plugin.Location)

	plugin.reader, err = plugin.ensureReader()
	if err != nil {
		return err
	}

	plugin.writer, err = bluge.OpenWriter(plugin.cfg)
	if err != nil {
		return err
	}

	go func() {
		for range plugin.readerTicker.C {
			_, err := plugin.ensureReader()
			if err != nil {
				deps.Logger().Error("failed to reopen reader", slog.Any("error", err))
			}
		}
	}()

	return plugin.SearchView.Setup(ctx, deps)
}

func (plugin *Search) ensureReader() (*bluge.Reader, error) {
	plugin.mx.Lock()
	defer plugin.mx.Unlock()

	if plugin.reader != nil {
		return plugin.reader, nil
	}

	newReader, err := bluge.OpenReader(plugin.cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errFailedOpeningBlugeReader, err)
	}

	plugin.reader = newReader

	return plugin.reader, nil
}

func (plugin *Search) Shutdown(_ context.Context) error {
	plugin.readerTicker.Stop()

	return errors.Join(
		plugin.writer.Close(),
		plugin.reader.Close(),
	)
}

func executeRequest[T any](
	ctx context.Context,
	reader *bluge.Reader,
	request bluge.SearchRequest,
	visitorMaker func(match *search.DocumentMatch, value *T) segment.StoredFieldVisitor,
) ([]T, error) {
	documentMatchIterator, err := reader.Search(ctx, request)
	if err != nil {
		return nil, err
	}

	match, err := documentMatchIterator.Next()

	if err != nil {
		return nil, err
	}

	results := make([]T, 0, match.Size())

	for match != nil {
		var result T
		visitor := visitorMaker(match, &result)

		err = match.VisitStoredFields(visitor)

		if err != nil {
			return nil, err
		}

		results = append(results, result)
		match, err = documentMatchIterator.Next()

		if err != nil {
			return results, err
		}
	}

	return results, nil
}
