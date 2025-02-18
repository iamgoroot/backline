package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
)

var (
	errDownloadFailed          = errors.New("failed to download")
	errMissingLocationMetadata = errors.New("location metadata is nil")
)

func (d *Discovery) TryDownload(
	ctx context.Context, deps core.Dependencies, locMeta *model.LocationMetadata, ref string,
) (string, error) {
	if locMeta == nil {
		return "", errMissingLocationMetadata
	}

	if locMeta.DownloadedBy == "github-discovery" && !strings.HasPrefix(ref, "http") {
		cfg, ok := locMeta.AdditionalInfo.(*Config)
		if !ok {
			deps.Logger().Error("invalid github location metadata")
		}

		ghGetter := &githubContentGetter{client: getClient(ctx, cfg), cfg: cfg}

		path := resolveRelativePath(locMeta, ref)

		content, err := getContent[fileContentQuery](ghGetter, ctx, path)
		if err != nil {
			deps.Logger().Error("failed to get content", slog.Any("err", err))
			return "", err
		}

		if content.Repository.Object.Blob.IsBinary {
			content, err := downloadBinaryFile(ctx, ghGetter.cfg.Owner, ghGetter.cfg.Repo, path, ghGetter.cfg.AccessToken)
			if err != nil {
				deps.Logger().Error("failed to download binary file", slog.Any("err", err))
				return "", err
			}

			return string(content), nil
		}

		return content.Repository.Object.Blob.Text, nil
	}

	return d.tryDownloadAbsoluteLink(ctx, ref)
}

func resolveRelativePath(locationMeta *model.LocationMetadata, ref string) string {
	path := locationMeta.Location
	lastPartIndex := strings.LastIndexByte(path, '/')

	if lastPartIndex == -1 {
		path = ""
	} else {
		path = path[:lastPartIndex]
	}

	pathItem := strings.TrimLeft(ref, "./")

	return strings.TrimLeft(fmt.Sprintf("%s/%s", path, pathItem), "/")
}

func (d *Discovery) tryDownloadAbsoluteLink(ctx context.Context, ref string) (string, error) {
	refData, err := parseGithubURL(ref)
	if err != nil {
		return "", err
	}

	for _, cfg := range d.cfgs {
		if refData.Host == cfg.Host || refData.Host == "github.com" && cfg.Host == "" {
			// use link data with access token from matching config
			ghConfig := &Config{
				Host:        refData.Host,
				Owner:       refData.Owner,
				Repo:        refData.Repo,
				Branch:      refData.Branch,
				AccessToken: cfg.AccessToken,
			}
			gh := &githubContentGetter{client: getClient(ctx, ghConfig), cfg: ghConfig}

			content, err := getContent[fileContentQuery](gh, ctx, refData.Path)
			if err == nil {
				txt := content.Repository.Object.Blob.Text
				if txt != "" {
					return txt, err
				}
			}
		}
	}

	return "", fmt.Errorf("%w: %s", errDownloadFailed, ref)
}
