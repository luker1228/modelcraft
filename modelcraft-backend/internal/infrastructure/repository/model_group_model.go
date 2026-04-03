package repository

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"time"
)

// ModelGroupPO is the persistence object for model_groups table.
type ModelGroupPO struct {
	ID           string
	OrgName      string
	ProjectSlug  string
	Name         string
	DisplayOrder string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName returns the table name for model groups.
func (ModelGroupPO) TableName() string {
	return "model_groups"
}

// ToModelGroup converts the persistence object to the domain entity.
func (p *ModelGroupPO) ToModelGroup() *modeldesign.ModelGroup {
	return &modeldesign.ModelGroup{
		ID: p.ID,
		ProjectScope: project.ProjectScope{
			OrgName:     p.OrgName,
			ProjectSlug: p.ProjectSlug,
		},
		Name:         p.Name,
		DisplayOrder: p.DisplayOrder,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// FromModelGroup populates the persistence object from the domain entity.
func (p *ModelGroupPO) FromModelGroup(g *modeldesign.ModelGroup) {
	p.ID = g.ID
	p.OrgName = g.OrgName
	p.ProjectSlug = g.ProjectSlug
	p.Name = g.Name
	p.DisplayOrder = g.DisplayOrder
	p.CreatedAt = g.CreatedAt
	p.UpdatedAt = g.UpdatedAt
}

// NewModelGroupPOFromDomain creates a ModelGroupPO from a domain ModelGroup.
func NewModelGroupPOFromDomain(g *modeldesign.ModelGroup) *ModelGroupPO {
	po := &ModelGroupPO{}
	po.FromModelGroup(g)
	return po
}
