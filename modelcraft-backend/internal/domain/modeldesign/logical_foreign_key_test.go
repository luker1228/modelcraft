package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalForeignKey_Validate_HappyPath(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		OrgName:      "test-org",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
		IsDeletable:  true,
	}
	assert.NoError(t, lf.Validate())
}

func TestLogicalForeignKey_Validate_RefModelIDOrTableRequired(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		OrgName:      "test-org",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RefModelID or RefTableName")

	lf.RefTableName = "users"
	assert.NoError(t, lf.Validate())
}

func TestLogicalForeignKey_Validate_EmptyID(t *testing.T) {
	lf := &LogicalForeignKey{
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestLogicalForeignKey_IsNormal(t *testing.T) {
	lf := &LogicalForeignKey{Direction: DirectionNormal}
	assert.True(t, lf.IsNormal())
	assert.False(t, lf.IsReverse())
}

func TestLogicalForeignKey_IsReverse(t *testing.T) {
	lf := &LogicalForeignKey{Direction: DirectionReverse}
	assert.True(t, lf.IsReverse())
	assert.False(t, lf.IsNormal())
}
