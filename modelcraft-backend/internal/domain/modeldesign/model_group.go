package modeldesign

import (
	"modelcraft/internal/domain/project"
	"regexp"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// UngroupedGroupID is the sentinel ID for the virtual ungrouped group.
// It is never stored in the database.
const UngroupedGroupID = "__ungrouped__"

// UngroupedGroupName is the display name for the virtual ungrouped group.
const UngroupedGroupName = "ungrouped"

// groupNamePattern matches valid group names: lowercase letters, digits, underscores; must start with a letter.
var groupNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// groupNameMaxLen is the maximum allowed length for a group name.
const groupNameMaxLen = 64

// ValidateGroupName validates a group name against the required pattern and length constraints.
// Returns an error if the name is empty, exceeds max length, matches the reserved "ungrouped" name,
// or does not match the pattern ^[a-z][a-z0-9_]*$.
func ValidateGroupName(name string) error {
	if name == "" {
		return bizerrors.NewValidationError("group name cannot be empty")
	}
	if len(name) > groupNameMaxLen {
		return bizerrors.NewValidationError("group name cannot exceed %d characters", groupNameMaxLen)
	}
	if name == UngroupedGroupName {
		return bizerrors.NewValidationError("group name %q is reserved", UngroupedGroupName)
	}
	if !groupNamePattern.MatchString(name) {
		return bizerrors.NewValidationError(
			"group name %q is invalid: must start with a letter and contain only "+
				"lowercase letters, digits, and underscores",
			name,
		)
	}
	return nil
}

// ModelGroup represents a named group that organizes models within a project.
// Models with no group assignment belong to the virtual ungrouped group.
type ModelGroup struct {
	ID                   string
	project.ProjectScope // 嵌入项目作用域，包含 OrgName 和 ProjectSlug
	Name                 string
	DisplayOrder         string
	Models               []*DataModel
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// IsVirtual reports whether this group is the virtual ungrouped sentinel.
// Virtual groups are synthesized by the application layer and never stored in the database.
func (g *ModelGroup) IsVirtual() bool {
	return g.ID == UngroupedGroupID
}

// NewUngroupedGroup returns the virtual ungrouped group sentinel.
// It has a well-known ID and name, and is always appended last in group listings.
func NewUngroupedGroup() *ModelGroup {
	return &ModelGroup{
		ID:   UngroupedGroupID,
		Name: UngroupedGroupName,
	}
}
