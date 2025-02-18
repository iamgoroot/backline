package github

import (
	"github.com/oriser/regroup"
)

var urlPattern = regroup.MustCompile(
	`^(?:https?:\/\/)?(?P<host>[^\/]+)\/(?P<owner>[^\/]+)\/(?P<repo>[^\/]+)\/blob\/(?P<branch>[^\/]+)\/(?P<path>.*)$`,
)

type githubData struct {
	Host   string `regroup:"host"`
	Owner  string `regroup:"owner"`
	Repo   string `regroup:"repo"`
	Branch string `regroup:"branch"`
	Path   string `regroup:"path"`
}

func parseGithubURL(location string) (githubData, error) {
	var data githubData
	err := urlPattern.MatchToTarget(location, &data)

	return data, err
}
