package modeldesign

// SchemaIssueType represents the type of schema drift issue detected
type SchemaIssueType string

const (
	// TableMissing indicates the underlying table does not exist
	TableMissing SchemaIssueType = "TABLE_MISSING"
	// FieldMissing indicates a field is missing from the table
	FieldMissing SchemaIssueType = "FIELD_MISSING"
	// FieldTypeMismatch indicates a field's type differs from the model definition
	FieldTypeMismatch SchemaIssueType = "FIELD_TYPE_MISMATCH"
	// FieldConstraintMismatch indicates a field's constraints differ from the model definition
	FieldConstraintMismatch SchemaIssueType = "FIELD_CONSTRAINT_MISMATCH"
	// DatabaseMissing indicates the database does not exist in the cluster
	DatabaseMissing SchemaIssueType = "DATABASE_MISSING"
	// ClusterNotFound indicates the cluster does not exist
	ClusterNotFound SchemaIssueType = "CLUSTER_NOT_FOUND"
	// FieldHasDependencies indicates a field cannot be deleted due to dependencies
	FieldHasDependencies SchemaIssueType = "FIELD_HAS_DEPENDENCIES"
)

// SchemaIssue represents a detected schema drift issue
type SchemaIssue struct {
	Type        SchemaIssueType
	Description string
	TableName   string
	FieldName   string
	Details     map[string]interface{}
}

// NewSchemaIssue creates a new schema issue
func NewSchemaIssue(
	issueType SchemaIssueType,
	description, tableName, fieldName string,
	details map[string]interface{},
) SchemaIssue {
	if details == nil {
		details = make(map[string]interface{})
	}
	return SchemaIssue{
		Type:        issueType,
		Description: description,
		TableName:   tableName,
		FieldName:   fieldName,
		Details:     details,
	}
}

// RepairMode represents the mode of repair operation
type RepairMode string

const (
	// DryRun detects issues without applying changes
	DryRun RepairMode = "DRY_RUN"
	// Additive adds missing resources only
	Additive RepairMode = "ADDITIVE"
	// FullSync adds missing and optionally removes extra resources
	FullSync RepairMode = "FULL_SYNC"
)

// HealthStatus represents the health status of a model
type HealthStatus string

const (
	// Healthy indicates no issues detected
	Healthy HealthStatus = "HEALTHY"
	// NeedsRepair indicates minor issues (missing fields only)
	NeedsRepair HealthStatus = "NEEDS_REPAIR"
	// Broken indicates major issues (missing table or cluster/database not found)
	Broken HealthStatus = "BROKEN"
)

// RepairResult represents the result of a repair operation
type RepairResult struct {
	Model              *DataModel
	ChangesApplied     bool
	DetectedIssues     []SchemaIssue
	ExecutedDDL        []string
	HealthStatusBefore HealthStatus
	HealthStatusAfter  HealthStatus
	ExtraFieldsRemoved []string
	FieldsAdded        []string
}
