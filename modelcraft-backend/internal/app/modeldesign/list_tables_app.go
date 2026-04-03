package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
)

// ListTablesResult holds paginated result for list tables
type ListTablesResult struct {
	Tables     []string
	TotalCount int
}

// ListTables 列出指定数据库中的所有基础表（BASE TABLE），可选排除已存在模型的表，支持分页
func (s *ReverseEngineerAppService) ListTables(
	ctx context.Context,
	orgName, projectSlug, databaseName string,
	excludeExisting bool,
	limit, offset int,
) (*ListTablesResult, error) {
	// 1. 获取数据库连接
	sqlDB, err := s.clusterManager.GetConnectionWithDatabase(ctx, orgName, projectSlug, databaseName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get database connection for %s", databaseName)
	}

	// 2. 查询 INFORMATION_SCHEMA.TABLES 获取所有基础表
	rows, err := sqlDB.QueryContext(
		ctx,
		"SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES"+
			" WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME",
		databaseName,
	)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to query tables for database %s", databaseName)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, bizerrors.Wrapf(err, "failed to scan table name")
		}
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return nil, bizerrors.Wrapf(err, "error iterating table rows")
	}

	if excludeExisting {
		tables, err = s.filterExistingTables(ctx, orgName, projectSlug, databaseName, tables)
		if err != nil {
			return nil, err
		}
	}

	totalCount := len(tables)

	// 4. 应用分页
	if limit > 0 {
		if offset >= len(tables) {
			return &ListTablesResult{Tables: []string{}, TotalCount: totalCount}, nil
		}
		end := offset + limit
		if end > len(tables) {
			end = len(tables)
		}
		tables = tables[offset:end]
	}

	return &ListTablesResult{Tables: tables, TotalCount: totalCount}, nil
}

// getExistingModelNames 查询指定项目+数据库下所有已存在的模型名
func (s *ReverseEngineerAppService) getExistingModelNames(
	ctx context.Context,
	orgName, projectSlug, databaseName string,
) ([]string, error) {
	query := modeldesign.ModelQuery{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
		Page:         1,
		PageSize:     10000, // 足够大以获取所有模型
	}

	models, _, err := s.modelRepo.Query(ctx, query)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to query existing models")
	}

	names := make([]string, 0, len(models))
	for _, m := range models {
		names = append(names, m.ModelLocator.ModelName)
	}
	return names, nil
}

// filterExistingTables 过滤掉已存在模型的表
func (s *ReverseEngineerAppService) filterExistingTables(
	ctx context.Context,
	orgName, projectSlug, databaseName string,
	tables []string,
) ([]string, error) {
	existingNames, err := s.getExistingModelNames(ctx, orgName, projectSlug, databaseName)
	if err != nil {
		return nil, err
	}
	if len(existingNames) == 0 {
		return tables, nil
	}
	existingSet := make(map[string]struct{}, len(existingNames))
	for _, name := range existingNames {
		existingSet[name] = struct{}{}
	}
	filtered := make([]string, 0, len(tables))
	for _, t := range tables {
		if _, found := existingSet[normalizeModelName(t)]; !found {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}
