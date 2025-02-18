package model

import (
	"net/url"
	"strings"
)

type ListEntityReq struct {
	Kind             string `param:"kind"`
	Sort             string `query:"sort"`
	Offset           int    `query:"offset"`
	Limit            int    `query:"limit"`
	WithDependencies bool   `query:"withDependencies"`
	ShowOrphans      bool   `query:"showOrphans"`
}

type Entity struct {
	FullName         string
	LocationMetadata *LocationMetadata

	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type LocationMetadata struct {
	AdditionalInfo any
	Location       string
	DownloadedBy   string
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Tags        []string          `yaml:"tags"`
	Links       []Link            `yaml:"links"`
}

type Link struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
	Icon  string `yaml:"icon"`
	Type  string `yaml:"type"`
}

type Spec struct {
	Profile *Profile `yaml:"profile"`

	Type       string `yaml:"type"`
	Lifecycle  string `yaml:"lifecycle"`
	Owner      string `yaml:"owner"`
	System     string `yaml:"system"`
	Definition any    `yaml:"definition"`
	Parent     string `yaml:"parent"`

	ParentOf      []string `yaml:"parentOf,omitempty"`
	ChildOf       []string `yaml:"childOf,omitempty"`
	OwnedBy       []string `yaml:"ownedBy,omitempty"`
	OwnerOf       []string `yaml:"ownerOf,omitempty"`
	DependsOn     []string `yaml:"dependsOn,omitempty"`
	DependencyOf  []string `yaml:"dependencyOf,omitempty"`
	ProvidesAPI   []string `yaml:"providesApi,omitempty"`
	APIProvidedBy []string `yaml:"apiProvidedBy,omitempty"`
	ConsumesAPI   []string `yaml:"consumesApi,omitempty"`
	APIConsumedBy []string `yaml:"apiConsumedBy,omitempty"`
	MemberOf      []string `yaml:"memberOf,omitempty"`
	HasMember     []string `yaml:"hasMember,omitempty"`
	PartOf        []string `yaml:"partOf,omitempty"`
	HasPart       []string `yaml:"hasPart,omitempty"`
	Children      []string `yaml:"children,omitempty"`
}

type Profile struct {
	DisplayName string `yaml:"displayName"`
	Email       string `yaml:"email"`
	Picture     string `yaml:"picture"`
}

func ParseFullName(name string) (kind, namespace, shortname string) {
	separateIndex := strings.IndexRune(name, ':')
	if separateIndex > 0 {
		kind = name[:separateIndex]
		name = name[separateIndex+1:]
	}

	namespace, shortname = parseNamespaceName(name)

	return kind, namespace, shortname
}

func parseNamespaceName(name string) (namespace, shortname string) {
	separateIndex := strings.IndexRune(name, '/')
	if separateIndex < 0 {
		return "default", name
	}

	return name[:separateIndex], name[separateIndex+1:]
}

func (e *Entity) FullNamePathEscaped() string {
	return url.PathEscape(e.FullName)
}
