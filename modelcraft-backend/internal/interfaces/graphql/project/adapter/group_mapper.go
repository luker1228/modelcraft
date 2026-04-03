package adapter

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
)

// GroupMapper is the singleton group mapper.
var GroupMapper = &groupMapper{}

type groupMapper struct{}

// ConvertToGraphQLGroup converts a domain ModelGroup to a GraphQL ModelGroup.
func (m *groupMapper) ConvertToGraphQLGroup(g *modeldesign.ModelGroup) *generated.ModelGroup {
	if g == nil {
		return m.virtualUngrouped()
	}
	return &generated.ModelGroup{
		ID:           g.ID,
		Name:         g.Name,
		IsVirtual:    g.IsVirtual(),
		DisplayOrder: g.DisplayOrder,
		Models:       []*generated.Model{},
	}
}

// GroupPlaceholder returns the appropriate ModelGroup stub to embed in a Model.
// When groupID is nil, the model belongs to the virtual ungrouped group.
// When groupID is set, it returns a stub with only the ID populated;
// the full group can be resolved lazily if needed.
func (m *groupMapper) GroupPlaceholder(groupID *string) *generated.ModelGroup {
	if groupID == nil {
		return m.virtualUngrouped()
	}
	return &generated.ModelGroup{
		ID:           *groupID,
		Name:         "",
		IsVirtual:    false,
		DisplayOrder: "",
		Models:       []*generated.Model{},
	}
}

func (m *groupMapper) virtualUngrouped() *generated.ModelGroup {
	return &generated.ModelGroup{
		ID:           modeldesign.UngroupedGroupID,
		Name:         modeldesign.UngroupedGroupName,
		IsVirtual:    true,
		DisplayOrder: "",
		Models:       []*generated.Model{},
	}
}
