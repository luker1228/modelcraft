package projectgraphql

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// FieldSelectionChecker provides field selection checking functionality
type FieldSelectionChecker struct{}

// NewFieldSelectionChecker creates a new field selection checker
func NewFieldSelectionChecker() *FieldSelectionChecker {
	return &FieldSelectionChecker{}
}

// IsFieldSelected checks if a specified field is selected in the GraphQL query.
// fieldPath supports dot-separated nested paths like "model", "model.fields", "edges.node.model".
func (f *FieldSelectionChecker) IsFieldSelected(ctx context.Context, fieldPath string) bool {
	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil {
		return false
	}

	operationCtx := graphql.GetOperationContext(ctx)
	if operationCtx == nil {
		return false
	}

	fields := graphql.CollectFields(operationCtx, fieldCtx.Field.Selections, nil)
	return f.containsField(operationCtx, fields, fieldPath)
}

// containsField checks if the field collection contains the specified field path.
// Supports dot-separated paths by recursing into nested selections.
func (f *FieldSelectionChecker) containsField(
	opCtx *graphql.OperationContext, fields []graphql.CollectedField, fieldPath string,
) bool {
	head, tail, hasNested := strings.Cut(fieldPath, ".")
	for _, field := range fields {
		if field.Name == head {
			if !hasNested {
				return true
			}
			nested := graphql.CollectFields(opCtx, field.Selections, nil)
			return f.containsField(opCtx, nested, tail)
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
