package modeldesign

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modelcraft/pkg/ddlfactory"
)

// ColumnDefinition represents a database column definition
type ColumnDefinition struct {
	Name          string
	DataType      string
	Length        int64
	Precision     int
	Scale         int
	Nullable      bool
	AutoIncrement bool
	DefaultValue  *string
	Comment       string
}

// TableDefinition represents a database table definition
type TableDefinition struct {
	TableName   string
	Columns     []ColumnDefinition
	PrimaryKeys []string
	Indexes     []IndexDefinition
	Comment     string
}

// IndexDefinition represents a database index definition
type IndexDefinition struct {
	Name    string
	Columns []string
	Unique  bool
	Primary bool
}

// RepairRequest represents a repair operation request
type RepairRequest struct {
	ProjectID         string
	ModelID           string
	Mode              RepairMode
	DeleteExtraFields bool
}

// SchemaComparisonService compares model definitions with actual database schema
type SchemaComparisonService interface {
	// CompareSchema compares a model's definition with the actual database schema
	// Returns detected issues and health status
	CompareSchema(
		ctx context.Context,
		model *DataModel,
		projectID string,
		clusterManager ClusterConnectionManager,
	) ([]SchemaIssue, HealthStatus, error)

	// GetTableDefinition retrieves the actual table definition from the database
	GetTableDefinition(
		ctx context.Context,
		orgName string,
		projectSlug string,
		databaseName string,
		tableName string,
		clusterManager ClusterConnectionManager,
	) (*TableDefinition, error)
}

// ClusterConnectionManager manages database cluster connections
type ClusterConnectionManager interface {
	GetConnectionWithDatabase(
		ctx context.Context,
		orgName string,
		projectSlug string,
		databaseName string,
	) (*sql.DB, error)
}

// MySQLSchemaComparisonService is the MySQL implementation of SchemaComparisonService
type MySQLSchemaComparisonService struct {
	typeMapper TypeMapper
}

// NewMySQLSchemaComparisonService creates a new MySQL schema comparison service
func NewMySQLSchemaComparisonService(typeMapper TypeMapper) SchemaComparisonService {
	return &MySQLSchemaComparisonService{
		typeMapper: typeMapper,
	}
}

// CompareSchema compares model definition with actual database schema
func (s *MySQLSchemaComparisonService) CompareSchema(
	ctx context.Context,
	model *DataModel,
	projectID string,
	clusterManager ClusterConnectionManager,
) ([]SchemaIssue, HealthStatus, error) {
	// TODO: Get orgName from context when needed
	orgName := "" // FIXME: Get from context

	conn, issues, status := s.getDatabaseConnection(ctx, model, orgName, clusterManager)
	if status != Healthy {
		return issues, status, nil
	}
	defer conn.Close()

	tableDef, issues := s.getTableDefinition(ctx, conn, model, issues)
	if len(issues) > 0 && issues[0].Type == TableMissing {
		return issues, Broken, nil
	}

	issues = s.compareModelFields(model, tableDef, issues)
	status = s.determineHealthStatus(issues)

	return issues, status, nil
}

// getDatabaseConnection gets database connection and handles connection errors
func (s *MySQLSchemaComparisonService) getDatabaseConnection(
	ctx context.Context,
	model *DataModel,
	orgName string,
	clusterManager ClusterConnectionManager,
) (*sql.DB, []SchemaIssue, HealthStatus) {
	conn, err := clusterManager.GetConnectionWithDatabase(
		ctx,
		orgName,
		model.ProjectSlug,
		model.DatabaseName,
	)
	if err != nil {
		issues, status := s.handleConnectionError(ctx, model, orgName, clusterManager, err)
		return nil, issues, status
	}
	return conn, nil, Healthy
}

// handleConnectionError determines if it's cluster or database issue
func (s *MySQLSchemaComparisonService) handleConnectionError(
	ctx context.Context,
	model *DataModel,
	orgName string,
	clusterManager ClusterConnectionManager,
	err error,
) ([]SchemaIssue, HealthStatus) {
	testConn, clusterErr := clusterManager.GetConnectionWithDatabase(
		ctx,
		orgName,
		model.ProjectSlug,
		"information_schema",
	)
	if testConn != nil {
		testConn.Close()
	}

	if clusterErr != nil {
		return []SchemaIssue{NewSchemaIssue(
			ClusterNotFound,
			"Cluster does not exist",
			model.ModelName,
			"",
			map[string]interface{}{"projectSlug": model.ProjectSlug, "error": clusterErr.Error()},
		)}, Broken
	}

	return []SchemaIssue{NewSchemaIssue(
		DatabaseMissing,
		"Database does not exist or cannot be accessed",
		model.ModelName,
		"",
		map[string]interface{}{"databaseName": model.DatabaseName, "error": err.Error()},
	)}, Broken
}

// getTableDefinition introspects the table and returns issues if missing
func (s *MySQLSchemaComparisonService) getTableDefinition(
	ctx context.Context,
	conn *sql.DB,
	model *DataModel,
	issues []SchemaIssue,
) (*TableDefinition, []SchemaIssue) {
	tableDef, err := s.introspectTable(ctx, conn, model.ModelName, model.DatabaseName)
	if err != nil {
		return nil, append(issues, NewSchemaIssue(
			TableMissing,
			"Table does not exist in the database",
			model.ModelName,
			"",
			map[string]interface{}{
				"projectSlug":  model.ProjectSlug,
				"databaseName": model.DatabaseName,
				"error":        err.Error(),
			},
		))
	}
	return tableDef, issues
}

// compareModelFields compares model fields with table columns
func (s *MySQLSchemaComparisonService) compareModelFields(
	model *DataModel,
	tableDef *TableDefinition,
	issues []SchemaIssue,
) []SchemaIssue {
	for _, field := range model.Fields {
		if field.IsEnumLabelField() {
			continue
		}

		col := s.findColumn(tableDef, field.Name)
		if col == nil {
			issues = append(issues, NewSchemaIssue(
				FieldMissing,
				"Field is missing from the table",
				model.ModelName,
				field.Name,
				map[string]interface{}{"expectedType": s.getExpectedType(field)},
			))
			continue
		}

		issues = s.checkFieldTypeMismatch(model, field, col, issues)
		issues = s.checkFieldConstraint(model, field, col, issues)
	}
	return issues
}

// checkFieldTypeMismatch checks if field type matches column type
func (s *MySQLSchemaComparisonService) checkFieldTypeMismatch(
	model *DataModel,
	field *FieldDefinition,
	col *ColumnDefinition,
	issues []SchemaIssue,
) []SchemaIssue {
	if !s.typesMatch(field, col, nil) {
		return append(issues, NewSchemaIssue(
			FieldTypeMismatch,
			"Field type does not match",
			model.ModelName,
			field.Name,
			map[string]interface{}{
				"expectedType": s.getExpectedType(field),
				"actualType":   col.DataType,
			},
		))
	}
	return issues
}

// checkFieldConstraint checks if field constraints match
func (s *MySQLSchemaComparisonService) checkFieldConstraint(
	model *DataModel,
	field *FieldDefinition,
	col *ColumnDefinition,
	issues []SchemaIssue,
) []SchemaIssue {
	if err := s.checkConstraintMismatch(field, col); err != nil {
		return append(issues, NewSchemaIssue(
			FieldConstraintMismatch,
			"Field constraint does not match",
			model.ModelName,
			field.Name,
			map[string]interface{}{
				"expectedNullable": !field.NonNull,
				"actualNullable":   col.Nullable,
			},
		))
	}
	return issues
}

// determineHealthStatus determines health status based on issues
func (s *MySQLSchemaComparisonService) determineHealthStatus(issues []SchemaIssue) HealthStatus {
	if len(issues) == 0 {
		return Healthy
	}

	for _, issue := range issues {
		if issue.Type == TableMissing || issue.Type == DatabaseMissing || issue.Type == ClusterNotFound {
			return Broken
		}
	}
	return NeedsRepair
}

// GetTableDefinition retrieves table definition from database
func (s *MySQLSchemaComparisonService) GetTableDefinition(
	ctx context.Context,
	orgName, projectSlug, databaseName, tableName string,
	clusterManager ClusterConnectionManager,
) (*TableDefinition, error) {
	conn, err := clusterManager.GetConnectionWithDatabase(ctx, orgName, projectSlug, databaseName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return s.introspectTable(ctx, conn, tableName, databaseName)
}

// introspectTable queries the schema of a table
func (s *MySQLSchemaComparisonService) introspectTable(
	ctx context.Context,
	conn *sql.DB,
	tableName,
	databaseName string,
) (*TableDefinition, error) {
	tableDef := &TableDefinition{
		TableName:   tableName,
		Columns:     []ColumnDefinition{},
		PrimaryKeys: []string{},
		Indexes:     []IndexDefinition{},
	}

	// Query columns
	columnsQuery := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			CASE
				WHEN CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN CHARACTER_MAXIMUM_LENGTH
				WHEN NUMERIC_PRECISION IS NOT NULL AND DATA_TYPE IN
					('decimal', 'numeric') THEN CONV(NUMERIC_PRECISION, 10, 10)
				ELSE NULL
			END AS COLUMN_LENGTH,
			COALESCE(NUMERIC_PRECISION, 0) AS NUMERIC_PRECISION,
			COALESCE(NUMERIC_SCALE, 0) AS NUMERIC_SCALE,
			IS_NULLABLE = 'YES' AS IS_NULLABLE,
			COLUMN_DEFAULT,
			EXTRA LIKE '%auto_increment%' AS AUTO_INCREMENT,
			COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := conn.QueryContext(ctx, columnsQuery, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnDefinition
		var length sql.NullInt64
		var defaultVal sql.NullString

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&length,
			&col.Precision,
			&col.Scale,
			&col.Nullable,
			&defaultVal,
			&col.AutoIncrement,
			&col.Comment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col.Length = length.Int64
		if defaultVal.Valid {
			col.DefaultValue = &defaultVal.String
		}

		tableDef.Columns = append(tableDef.Columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	if len(tableDef.Columns) == 0 {
		return nil, fmt.Errorf("table %s not found or has no columns", tableName)
	}

	// Query primary keys
	keysQuery := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION
	`

	keyRows, err := conn.QueryContext(ctx, keysQuery, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary keys: %w", err)
	}
	defer keyRows.Close()

	for keyRows.Next() {
		var key string
		if err := keyRows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan primary key: %w", err)
		}
		tableDef.PrimaryKeys = append(tableDef.PrimaryKeys, key)
	}

	return tableDef, nil
}

// findColumn finds a column by name in the table definition
func (s *MySQLSchemaComparisonService) findColumn(table *TableDefinition, name string) *ColumnDefinition {
	for i := range table.Columns {
		if table.Columns[i].Name == name {
			return &table.Columns[i]
		}
	}
	return nil
}

// typesMatch checks if field type matches column type
func (s *MySQLSchemaComparisonService) typesMatch(
	field *FieldDefinition,
	col *ColumnDefinition,
	table *TableDefinition,
) bool {
	expectedTypeBase := parseMySQLType(col.DataType)

	if s.typeMapper == nil {
		return false
	}

	mappedType, err := s.typeMapper.MapToMySQL(field)
	if err != nil {
		return false
	}

	// Compare base types first (VARCHAR vs TEXT is compatible)
	mappedBase := parseMySQLType(mappedType)
	if mappedBase == expectedTypeBase {
		return true
	}
	// For string types, check compatibility
	if isStringType(mappedBase) && isStringType(expectedTypeBase) {
		return true
	}

	return false
}

// getExpectedType gets the expected MySQL type for a field
func (s *MySQLSchemaComparisonService) getExpectedType(field *FieldDefinition) string {
	if s.typeMapper != nil {
		if typ, err := s.typeMapper.MapToMySQL(field); err == nil {
			return typ
		}
	}
	return "UNKNOWN"
}

// parseMySQLType parses the base type from a MySQL type string
func parseMySQLType(mysqlType string) string {
	// Simple parser to extract base type
	for i, c := range mysqlType {
		if c == '(' || c == ' ' {
			return mysqlType[:i]
		}
	}
	return mysqlType
}

// isStringType checks if a type is a string type
func isStringType(baseType string) bool {
	switch baseType {
	case string(ddlfactory.CHAR), string(ddlfactory.VARCHAR), string(ddlfactory.TEXT),
		"TINYTEXT", "MEDIUMTEXT", "LONGTEXT", string(ddlfactory.JSON):
		return true
	}
	return false
}

// checkConstraintMismatch checks if field constraints match
func (s *MySQLSchemaComparisonService) checkConstraintMismatch(field *FieldDefinition, col *ColumnDefinition) error {
	if field.NonNull && col.Nullable {
		return errors.New("field should be NOT NULL but column is nullable")
	}
	// Column being NOT NULL when field is nullable is acceptable (more restrictive)
	return nil
}

// IsFieldReferencedByRelation checks if a field is referenced by any relation
func IsFieldReferencedByRelation(model *DataModel, fieldName string) bool {
	for _, field := range model.Fields {
		if field.ParentRelationID != nil && field.Name == fieldName {
			return true
		}
	}
	return false
}
