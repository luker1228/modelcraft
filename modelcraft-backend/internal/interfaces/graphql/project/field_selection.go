package projectgraphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

// FieldSelectionChecker provides field selection checking functionality
type FieldSelectionChecker struct{}

// NewFieldSelectionChecker creates a new field selection checker
func NewFieldSelectionChecker() *FieldSelectionChecker {
	return &FieldSelectionChecker{}
}

// IsFieldSelected checks if a specified field is selected in the GraphQL query
// fieldPath supports nested field paths like "model", "model.fields", "edges.node.model" etc.
func (f *FieldSelectionChecker) IsFieldSelected(ctx context.Context, fieldPath string) bool {
	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil {
		return false
	}

	// Get the current operation context
	operationCtx := graphql.GetOperationContext(ctx)
	if operationCtx == nil {
		return false
	}

	// Collect all child fields of the current field
	fields := graphql.CollectFields(operationCtx, fieldCtx.Field.Selections, nil)

	// Check if it contains the specified field path
	return f.containsField(fields, fieldPath)
}

// containsField recursively checks if the field collection contains the specified field
func (f *FieldSelectionChecker) containsField(fields []graphql.CollectedField, fieldPath string) bool {
	// Simple field name check
	for _, field := range fields {
		if field.Name == fieldPath {
			return true
		}
	}
	return false
}

// IsAnyFieldSelected checks if any field from the specified list is selected
func (f *FieldSelectionChecker) IsAnyFieldSelected(ctx context.Context, fieldPaths []string) bool {
	for _, fieldPath := range fieldPaths {
		if f.IsFieldSelected(ctx, fieldPath) {
			return true
		}
	}
	return false
}

// GetSelectedFields gets all selected field names in the current query
func (f *FieldSelectionChecker) GetSelectedFields(ctx context.Context) []string {
	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil {
		return nil
	}

	operationCtx := graphql.GetOperationContext(ctx)
	if operationCtx == nil {
		return nil
	}

	fields := graphql.CollectFields(operationCtx, fieldCtx.Field.Selections, nil)

	selectedFields := make([]string, 0, len(fields))
	for _, field := range fields {
		selectedFields = append(selectedFields, field.Name)
	}

	return selectedFields
}
