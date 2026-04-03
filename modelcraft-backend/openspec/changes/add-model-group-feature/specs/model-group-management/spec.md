## ADDED Requirements

### Requirement: Create group within a project
The system SHALL allow users to create a named group within a project. The group name MUST match `^[a-z][a-z0-9_]*$` and MUST be unique within the project (case-sensitive). Maximum length is 64 characters. The new group SHALL be assigned a `display_order` value placing it at the tail of the existing groups.

#### Scenario: Successful group creation
- **WHEN** a user creates a group with name `payment` in a project that has no group named `payment`
- **THEN** a new group is created and returned with a valid ID, the given name, and a `display_order` placing it last

#### Scenario: Duplicate group name rejected
- **WHEN** a user creates a group with a name that already exists in the same project
- **THEN** the system SHALL return a `GroupAlreadyExists` error and no group is created

#### Scenario: Invalid group name rejected
- **WHEN** a user creates a group with a name that does not match `^[a-z][a-z0-9_]*$` (e.g., `_bad`, `2fast`, `My-Group`)
- **THEN** the system SHALL return an `InvalidGroupName` error and no group is created

#### Scenario: Group name at max length accepted
- **WHEN** a user creates a group with a name exactly 64 characters long matching the pattern
- **THEN** the group is created successfully

#### Scenario: Group name exceeding max length rejected
- **WHEN** a user creates a group with a name longer than 64 characters
- **THEN** the system SHALL return an `InvalidGroupName` error

---

### Requirement: Rename a group
The system SHALL allow users to rename an existing group. The new name MUST satisfy the same validation rules as creation. The new name MUST be unique within the project. All models referencing this group SHALL immediately reflect the new name without additional writes to the models table.

#### Scenario: Successful rename
- **WHEN** a user renames group `payment` to `payments` and no group named `payments` exists in the project
- **THEN** the group name is updated and the group is returned with the new name

#### Scenario: Rename to conflicting name rejected
- **WHEN** a user renames a group to a name already used by another group in the same project
- **THEN** the system SHALL return a `GroupAlreadyExists` error and the group name is unchanged

#### Scenario: Rename to invalid name rejected
- **WHEN** a user renames a group to a name that does not match `^[a-z][a-z0-9_]*$`
- **THEN** the system SHALL return an `InvalidGroupName` error and the group name is unchanged

#### Scenario: Rename non-existent group
- **WHEN** a user attempts to rename a group ID that does not exist
- **THEN** the system SHALL return a `GroupNotFound` error

---

### Requirement: Delete a group
The system SHALL allow users to delete a group. Upon deletion, all models assigned to that group SHALL have their `group_id` set to `NULL` (moved to virtual ungrouped). The deletion and model reassignment SHALL occur atomically within a single transaction.

#### Scenario: Delete group with models
- **WHEN** a user deletes a group that contains one or more models
- **THEN** the group is deleted and all previously assigned models now belong to the virtual ungrouped group

#### Scenario: Delete empty group
- **WHEN** a user deletes a group that contains no models
- **THEN** the group is deleted successfully

#### Scenario: Delete non-existent group
- **WHEN** a user attempts to delete a group ID that does not exist
- **THEN** the system SHALL return a `GroupNotFound` error

---

### Requirement: Reorder groups within a project
The system SHALL allow users to change the display order of a group relative to other groups. The system SHALL use lexicographic fractional indexing to compute a new `display_order` value, requiring only a single row update per reorder operation.

#### Scenario: Move group to head
- **WHEN** a user reorders a group to position before all other groups (afterGroupId = null)
- **THEN** the group's `display_order` is updated to a value less than all other groups in the project

#### Scenario: Move group between two existing groups
- **WHEN** a user reorders a group to position after a specific group
- **THEN** the group's `display_order` is updated to a lexicographic midpoint between the target group and the next group

#### Scenario: Move group to tail
- **WHEN** a user reorders a group to position after the last group
- **THEN** the group's `display_order` is updated to a value greater than all other groups in the project

---

### Requirement: Assign a model to a group
The system SHALL allow users to assign a model to a group within the same project. A model SHALL belong to exactly one group at any time. Assigning a model to a group replaces any previous group assignment.

#### Scenario: Assign model to a group
- **WHEN** a user assigns a model to a group in the same project
- **THEN** the model's `group_id` is updated and the model's `group` field reflects the new group

#### Scenario: Move model to ungrouped
- **WHEN** a user assigns a model with `groupId = null`
- **THEN** the model's `group_id` is set to `NULL` and the model's `group` field returns the virtual ungrouped group

#### Scenario: Assign model to group in different project rejected
- **WHEN** a user attempts to assign a model to a group belonging to a different project
- **THEN** the system SHALL return a `GroupNotFound` error (group is not visible outside its project)

---

### Requirement: List groups in a project
The system SHALL return all groups in a project ordered by `display_order` ascending, followed by the virtual ungrouped group appended last. Each group SHALL include its associated models.

#### Scenario: List groups with models
- **WHEN** a user queries groups for a project
- **THEN** the system returns all real groups ordered by `display_order`, with each group containing its assigned models, and the virtual ungrouped group appended last containing models with no group assignment

#### Scenario: List groups in empty project
- **WHEN** a user queries groups for a project with no groups and no models
- **THEN** the system returns only the virtual ungrouped group with an empty models list

---

### Requirement: Virtual ungrouped group
The system SHALL provide a virtual "ungrouped" group for every project. This group SHALL never be stored in the database. It SHALL have a sentinel ID of `__ungrouped__`, name `ungrouped`, and `isVirtual = true`. It SHALL always appear last in group listings. It SHALL contain all models whose `group_id IS NULL`. It SHALL NOT be renameable, deleteable, or reorderable.

#### Scenario: Models without group appear in ungrouped
- **WHEN** a model has no group assignment (`group_id IS NULL`)
- **THEN** the model's `group` field returns the virtual ungrouped group with `id = "__ungrouped__"` and `isVirtual = true`

#### Scenario: Ungrouped always last in listing
- **WHEN** a user queries groups for a project
- **THEN** the virtual ungrouped group is always the last entry regardless of other groups' display_order values
