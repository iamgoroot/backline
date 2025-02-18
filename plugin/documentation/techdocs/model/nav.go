package model

type Nav struct {
	Name     string
	URL      string
	Children []*Nav
}
