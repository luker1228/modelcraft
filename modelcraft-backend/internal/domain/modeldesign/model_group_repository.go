package modeldesign

import "context"

// ModelGroupRepository defines the data access interface for model groups.
// Implementations must be safe for concurrent use.
type ModelGroupRepository interface {
	// Create persists a new group. Returns an error if the name already exists within the project.
	Create(ctx context.Context, group *ModelGroup) error

	// FindByID retrieves a group by its unique identifier.
	// Returns nil, nil if the group does not exist.
	FindByID(ctx context.Context, id string) (*ModelGroup, error)

	// FindByName retrieves a group by its name within a project scope.
	// Returns nil, nil if the group does not exist.
	FindByName(ctx context.Context, orgName, projectSlug, name string) (*ModelGroup, error)

	// ListByProject retrieves all groups in a project ordered by display_order ascending.
	ListByProject(ctx context.Context, orgName, projectSlug string) ([]*ModelGroup, error)

	// Update persists changes to an existing group (name, display_order).
	Update(ctx context.Context, group *ModelGroup) error

	// Delete removes a group by ID. Does not affect models; caller must handle model reassignment.
	Delete(ctx context.Context, id string) error

	// UpdateModelsGroup sets group_id = newGroupID on all models currently in groupID.
	// Pass nil for newGroupID to move models to ungrouped.
	UpdateModelsGroup(ctx context.Context, groupID string, newGroupID *string) error

	// GetTailDisplayOrder returns the largest display_order value among groups in the project,
	// or an empty string if no groups exist. Used to compute new tail position on create.
	GetTailDisplayOrder(ctx context.Context, orgName, projectSlug string) (string, error)
}
