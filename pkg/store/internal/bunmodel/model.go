package bunmodel

import (
	"time"

	"github.com/iamgoroot/backline/pkg/model"
)

type StoredEntity struct {
	OrphanedAt *time.Time
	model.Profile
	FullName   string `bun:",pk,unique"`
	UnkindName string
	APIVersion string
	Kind       string
	Spec
	model.Metadata
}

type Spec struct {
	Definition any `bun:"type:jsonb"`
	Type       string
	Lifecycle  string
	Owner      string
	System     string
	Relations  Relations `bun:"type:jsonb"`
}
type Relations struct {
	OwnedBy       []string `json:",omitempty"`
	OwnerOf       []string `json:",omitempty"`
	ProvidesAPI   []string `json:",omitempty"`
	APIProvidedBy []string `json:",omitempty"`
	ConsumesAPI   []string `json:",omitempty"`
	APIConsumedBy []string `json:",omitempty"`
	DependsOn     []string `json:",omitempty"`
	DependencyOf  []string `json:",omitempty"`
	ParentOf      []string `json:",omitempty"`
	ChildOf       []string `json:",omitempty"`
	MemberOf      []string `json:",omitempty"`
	HasMember     []string `json:",omitempty"`
	PartOf        []string `json:",omitempty"`
	HasPart       []string `json:",omitempty"`
	Children      []string `json:",omitempty"`
}

type EntityMapping struct {
	From        *StoredEntity `bun:"rel:belongs-to,join:from_id=full_name"`
	To          *StoredEntity `bun:"rel:belongs-to,join:to_id=full_name"`
	Relation    string        `bun:",pk,unique:unique_entity_mapping"`
	FromID      string        `bun:",pk,unique:unique_entity_mapping"`
	ToID        string        `bun:",pk,unique:unique_entity_mapping"`
	RefEntityID string
}
