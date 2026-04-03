package modeldesign

import (
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newBaseFieldForFKTest() *FieldDefinition {
	relateFkID := "lf-001"
	ft, _ := NewFieldFormat(FormatRelation)
	return &FieldDefinition{
		ModelID: "model-order",
		Name:    "user",
		Title:   "User",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			DatabaseName: "db",
			ModelName:    "order",
		},
		Type:       ft,
		RelateFKID: &relateFkID,
	}
}

// TestFieldDefinition_BelongsToAndRelate_MutuallyExclusive 验证 belongs_to_fk_id 和 relate_fk_id 互斥
func TestFieldDefinition_BelongsToAndRelate_MutuallyExclusive(t *testing.T) {
	belongsToID := "lf-001"
	relateFkID := "lf-002"
	ft, _ := NewFieldFormat(FormatString)
	fd := &FieldDefinition{
		ModelID: "model-order",
		Name:    "userId",
		Title:   "User ID",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			DatabaseName: "db",
			ModelName:    "order",
		},
		Type:          ft,
		BelongsToFKID: &belongsToID,
		RelateFKID:    &relateFkID,
	}
	err := fd.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestFieldDefinition_RelationFormat_RequiresRelateFkId 验证 RELATION 格式字段必须有 relate_fk_id
func TestFieldDefinition_RelationFormat_RequiresRelateFkId(t *testing.T) {
	ft, _ := NewFieldFormat(FormatRelation)
	fd := &FieldDefinition{
		ModelID: "model-order",
		Name:    "user",
		Title:   "User",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			DatabaseName: "db",
			ModelName:    "order",
		},
		Type:       ft,
		RelateFKID: nil, // 没有设置 relate_fk_id
	}
	err := fd.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RELATION format field must have relate_fk_id")
}

// TestFieldDefinition_RelationFormat_WithRelateFkId 验证 RELATION 格式字段有 relate_fk_id 时通过验证
func TestFieldDefinition_RelationFormat_WithRelateFkId(t *testing.T) {
	fd := newBaseFieldForFKTest()
	err := fd.Validate()
	assert.NoError(t, err)
}

// TestFieldDefinition_BelongsToFKID_OnlySet 验证仅设置 belongs_to_fk_id 时通过验证
func TestFieldDefinition_BelongsToFKID_OnlySet(t *testing.T) {
	belongsToID := "lf-001"
	ft, _ := NewFieldFormat(FormatString)
	fd := &FieldDefinition{
		ModelID: "model-order",
		Name:    "userId",
		Title:   "User ID",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			DatabaseName: "db",
			ModelName:    "order",
		},
		Type:          ft,
		BelongsToFKID: &belongsToID,
	}
	err := fd.Validate()
	assert.NoError(t, err)
}

// TestFieldDefinition_NeitherFKField 验证普通字段不设置 FK 字段时通过验证
func TestFieldDefinition_NeitherFKField(t *testing.T) {
	ft, _ := NewFieldFormat(FormatString)
	fd := &FieldDefinition{
		ModelID: "model-order",
		Name:    "name",
		Title:   "Name",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			DatabaseName: "db",
			ModelName:    "order",
		},
		Type: ft,
	}
	err := fd.Validate()
	assert.NoError(t, err)
}
