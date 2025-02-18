package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iamgoroot/backline/pkg/core"
	"github.com/iamgoroot/backline/pkg/model"
	discovery "github.com/iamgoroot/backline/plugin/discovery/fs"
	"github.com/shurcooL/githubv4"
)

type githubContentGetter struct {
	client *githubv4.Client
	cfg    *Config
}

func (gh *githubContentGetter) getLatestVersion(ctx context.Context) string {
	content, err := getContent[commitQuery](gh, ctx, "")
	if err != nil {
		return ""
	}

	return content.Repository.Object.CommitResourcePath
}

func (gh *githubContentGetter) processEntities(ctx context.Context, path string, depth int, register core.RegistrationFunc) error {
	content, err := getContent[contentQuery](gh, ctx, path)
	if err != nil {
		return err
	}

	var errs error

	contents := content.Repository.Object.Tree.Entries
	for _, content := range contents {
		if depth <= gh.cfg.Depth && content.Type == "tree" {
			err = gh.processEntities(ctx, content.Path, depth+1, register)
			errs = errors.Join(errs, err)
		}

		if content.Type == "blob" && (content.Extension == ".yaml" || content.Extension == ".yml") {
			location := &model.LocationMetadata{
				DownloadedBy:   "github-discovery",
				Location:       content.Path,
				AdditionalInfo: gh.cfg,
			}
			err = discovery.ReadSpecs(strings.NewReader(content.Object.Blob.Text), location, register)
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func getContent[Q any](ghGetter *githubContentGetter, ctx context.Context, path string) (*Q, error) {
	vars := map[string]interface{}{
		"repo":       githubv4.String(ghGetter.cfg.Repo),
		"owner":      githubv4.String(ghGetter.cfg.Owner),
		"expression": githubv4.String(fmt.Sprintf("%s:%s", ghGetter.cfg.Branch, path)),
	}

	var q Q
	err := ghGetter.client.Query(ctx, &q, vars)

	return &q, err
}
