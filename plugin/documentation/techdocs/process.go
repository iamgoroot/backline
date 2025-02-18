package techdocs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	"github.com/iamgoroot/backline/plugin/documentation/techdocs/internal/mkdocs"
	img64 "github.com/tenkoh/goldmark-img64"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
	"golang.org/x/sync/errgroup"
)

func (p *Plugin) ProcessEntity(ctx context.Context, _ core.Dependencies, entity *model.Entity) error {
	docsRoot, ok := entity.Metadata.Annotations["backstage.io/techdocs-ref"]
	if !ok {
		return nil
	}

	docsRoot = sanitizeDocRoot(docsRoot)

	discovery, docsIndexSrc := p.downloadDocIndex(ctx, entity.LocationMetadata, docsRoot)
	if docsIndexSrc != "" && discovery != nil {
		group := fmt.Sprintf("techdocs-items:%s", entity.FullName)
		docItemProcessor := docItemProcessor{
			store: func(key string, data string) error {
				return p.StoreKV().Set(ctx, group, key, data, 0)
			},
			download: func(ref string) (string, error) {
				return discovery.TryDownload(ctx, p.Dependencies, entity.LocationMetadata, getDocItemPath(docsRoot, ref))
			},
		}

		return p.processDocs(ctx, entity.FullName, docItemProcessor, []byte(docsIndexSrc))
	}

	return nil
}

func getDocItemPath(root, path string) string {
	pathPart := strings.TrimLeft(path, "./")
	path = fmt.Sprintf("%s/docs/%s", root, pathPart)

	return strings.TrimLeft(path, "./")
}
func sanitizeDocRoot(root string) string {
	root = strings.TrimPrefix(root, "dir:")
	root = strings.TrimPrefix(root, "url:")
	root = strings.TrimSuffix(root, "/")

	return strings.TrimLeft(root, "./")
}
func (p *Plugin) processDocs(ctx context.Context, entityName string, docItemProcessor docItemProcessor, docsIndexSrc []byte) error {
	indexWalkFunc, wait := docItemProcessor.createAsyncWalker(p.Parallelism)

	docsIndex, err := mkdocs.ParseDocIndex(docsIndexSrc, indexWalkFunc)
	if err != nil {
		p.Logger().Error("failed to parse docs index", slog.Any("err", err))
		return err
	}

	if err = wait(); err != nil {
		p.Logger().Error("failed to download docs", slog.Any("err", err))
	}

	err = p.StoreKV().Set(ctx, "techdocs-root", entityName, docsIndex, 0)

	return err
}

func (p *Plugin) downloadDocIndex(ctx context.Context, loc *model.LocationMetadata, root string) (loader core.Discovery, content string) {
	ref := fmt.Sprintf("%s/%s", root, "mkdocs.yml")
	for _, discovery := range p.Discoveries() {
		data, err := discovery.TryDownload(ctx, p.Dependencies, loc, ref)
		if err == nil && data != "" {
			return discovery, data
		}
	}

	return nil, ""
}

func createMarkdownConverter(imageDownloadFunc func(ref string) (string, error)) func([]byte, io.Writer) error {
	docFileReader := func(name string) ([]byte, error) {
		if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
			return []byte(name), nil
		}

		content, err := imageDownloadFunc(name)
		if err != nil {
			return nil, err
		}

		return []byte(content), nil
	}

	return func(input []byte, output io.Writer) error {
		return goldmark.New(
			goldmark.WithExtensions(extension.GFM, &frontmatter.Extender{}, img64.Img64),
			goldmark.WithRendererOptions(img64.WithFileReader(docFileReader)),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		).Convert(input, output)
	}
}

type docItemProcessor struct {
	download func(ref string) (string, error)
	store    func(string, string) error
}

func (w docItemProcessor) createAsyncWalker(paralelism int) (walk func(url string), wait func() error) {
	errGroup := errgroup.Group{}
	errGroup.SetLimit(paralelism)

	indexWalkFunc := func(url string) { //TODO: switch to aggregate errors
		errGroup.Go(w.process(url))
	}

	return indexWalkFunc, errGroup.Wait
}

var (
	errDownloadFailed   = errors.New("failed to download content")
	errConversionFailed = errors.New("failed to convert content")
	errStorageFailed    = errors.New("failed to store content")
)

func (w docItemProcessor) process(ref string) func() error {
	markdownConverterFunc := createMarkdownConverter(w.download)

	return func() error {
		// Download the content
		content, err := w.download(ref)
		if err != nil {
			return fmt.Errorf("%w: %s: %v", errDownloadFailed, ref, err)
		}

		result := bytes.Buffer{}

		// Convert the content
		err = markdownConverterFunc([]byte(content), &result)
		if err != nil {
			return fmt.Errorf("%w: %s: %v", errConversionFailed, ref, err)
		}

		// Store the result
		err = w.store(ref, result.String())
		if err != nil {
			return fmt.Errorf("%w: %s: %v", errStorageFailed, ref, err)
		}

		return nil
	}
}
