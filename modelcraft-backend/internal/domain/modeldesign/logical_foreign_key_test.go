package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogicalForeignKey_Validate_HappyPath(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
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

func TestLogicalForeignKey_Validate_EmptyPairID(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
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
	assert.Contains(t, err.Error(), "PairID cannot be empty")
}

func TestLogicalForeignKey_Validate_InvalidDirection(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    "invalid",
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "direction must be")
}

func TestLogicalForeignKey_Validate_EmptyModelID(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ModelID cannot be empty")
}

func TestLogicalForeignKey_Validate_EmptyRefModelID(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RefModelID cannot be empty")
}

func TestLogicalForeignKey_Validate_EmptySourceFields(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SourceFields cannot be empty")
}

func TestLogicalForeignKey_Validate_FieldCountMismatch(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId", "companyId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count mismatch")
}

func TestLogicalForeignKey_Validate_CompositeFK(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId", "companyId"},
		TargetFields: []string{"id", "companyId"},
	}
	assert.NoError(t, lf.Validate())
}

func TestLogicalForeignKey_Validate_EmptyModelName(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ModelName cannot be empty")
}

func TestLogicalForeignKey_Validate_EmptyRefModelName(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}
	err := lf.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RefModelName cannot be empty")
}

func TestLogicalForeignKey_Validate_ReverseDirection(t *testing.T) {
	lf := &LogicalForeignKey{
		ID:           "lf-002",
		PairID:       "pair-001",
		Direction:    DirectionReverse,
		ModelID:      "model-user",
		ModelName:    "User",
		RefModelID:   "model-order",
		RefModelName: "Order",
		SourceFields: []string{"id"},
		TargetFields: []string{"userId"},
	}
	assert.NoError(t, lf.Validate())
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
