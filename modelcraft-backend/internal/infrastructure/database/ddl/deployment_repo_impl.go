package ddl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"time"

	"github.com/google/uuid"
)

// DeploymentImpl 部署服务，负责将模型变更部署到客户数据库
type DeploymentImpl struct {
	clusterManager *repository.ClusterConnectionManager // 集群连接管理器
}

// DeployModelToRemoveFields 部署模型删除字段
func (d *DeploymentImpl) DeployModelToRemoveFields(
	ctx context.Context,
	model *modeldesign.DataModel,
	fieldKeys []string,
) error {
	return nil
}

// NewDeploymentService 创建部署服务
func NewDeploymentService(clusterManager *repository.ClusterConnectionManager) *DeploymentImpl {
	return &DeploymentImpl{
		clusterManager: clusterManager,
	}
}

// DeployModelToCreate 部署模型创建
func (d *DeploymentImpl) DeployModelToCreate(ctx context.Context, model *modeldesign.DataModel) error {
	converter := newDdlConverter(ctx)
	ddl, err := converter.GenerateCreateTableDDL(model)
	if err != nil {
		logfacade.GetLogger(ctx).Infof(ctx, "gen create table ddl fail: %v", logfacade.Err(err))
		return shared.NewRepositoryError(shared.ErrTypeSQLConvertion, "DDL generate fail")
	}

	repositoryErr := d.deployToClientDatabase(
		ctx,
		model.GetModelLocator(),
		ddl,
		modeldesign.ChangeCreateModel,
	)
	if repositoryErr != nil {
		return repositoryErr
	}
	return nil
}

// DeployModelToDrop 部署模型删除
func (d *DeploymentImpl) DeployModelToDrop(ctx context.Context, model *modeldesign.DataModel) error {
	converter := newDdlConverter(ctx)
	ddl, err := converter.GenerateDropTableDDL(model)
	if err != nil {
		return shared.NewRepositoryError(shared.ErrTypeSQLConvertion, "DDL generate fail")
	}

	repositoryErr := d.deployToClientDatabase(
		ctx,
		model.GetModelLocator(),
		ddl,
		modeldesign.ChangeDropModel,
	)
	if repositoryErr != nil {
		return repositoryErr
	}

	return nil
}

func (s *DeploymentImpl) deployToClientDatabase(
	ctx context.Context,
	locator *modeldesign.ModelLocator,
	ddlStatement string,
	changeType modeldesign.ModelChangeType,
) *shared.RepositoryError {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(
		ctx, "ddl=%s modelName=%s, changeType=%s, database=%s",
		ddlStatement, locator.ModelName, changeType, locator.DatabaseName,
	)

	// 获取集群连接
	// Get orgName from context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return shared.NewRepositoryError(
			shared.ErrTypeUnknown,
			fmt.Sprintf("获取组织名称失败: %v", err),
		)
	}
	conn, err := s.clusterManager.GetConnectionWithDatabase(
		ctx,
		orgName,
		locator.ProjectSlug,
		locator.DatabaseName,
	)
	if err != nil {
		return shared.NewRepositoryError(
			shared.ErrTypeConnection,
			fmt.Sprintf("获取数据库连接失败: %v", err),
		)
	}

	// 执行DDL语句
	// NOCA: yunding/go/sql-injection
	_, err = conn.ExecContext(ctx, ddlStatement)
	if err != nil {
		return shared.NewRepositoryError(
			shared.ErrTypeDDL,
			fmt.Sprintf("执行DDL失败: %s, 错误: %v", ddlStatement, err),
		)
	}
	return nil
}

// CheckTableExists checks whether the underlying database table for the given model already exists.
// Returns (true, nil) if the table exists, (false, nil) if it does not, or (false, err) on failure.
func (d *DeploymentImpl) CheckTableExists(ctx context.Context, model *modeldesign.DataModel) (bool, error) {
	logger := logfacade.GetLogger(ctx)

	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get organization name: %w", err)
	}

	conn, err := d.clusterManager.GetConnectionWithDatabase(
		ctx,
		orgName,
		model.ProjectSlug,
		model.DatabaseName,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get database connection: %w", err)
	}

	// NOCA: yunding/go/sql-injection
	var tableName string
	query := "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? LIMIT 1"
	row := conn.QueryRowContext(ctx, query, model.DatabaseName, model.ModelName)
	scanErr := row.Scan(&tableName)
	if scanErr != nil {
		// sql.ErrNoRows means the table does not exist
		if errors.Is(scanErr, sql.ErrNoRows) {
			logger.Infof(ctx, "table %s.%s does not exist", model.DatabaseName, model.ModelName)
			return false, nil
		}
		return false, fmt.Errorf("failed to check table existence: %w", scanErr)
	}

	logger.Infof(ctx, "table %s.%s already exists", model.DatabaseName, model.ModelName)
	return true, nil
}

func (s *DeploymentImpl) CreateDeploymentHistory(
	modelID string,
	changeType modeldesign.ModelChangeType,
	ddlStatements []string,
) *modeldesign.DeploymentHistory {
	return &modeldesign.DeploymentHistory{
		ID:            uuid.New().String(),
		ModelID:       modelID,
		ChangeType:    changeType,
		DDLStatements: ddlStatements,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}
}

// GetPendingDeployments 获取待部署的模型列表
func (s *DeploymentImpl) GetPendingDeployments() ([]string, error) {
	// 这里应该查询数据库获取部署状态为pending或failed的模型
	// 简化实现，返回空列表
	return []string{}, nil
}

// BatchRetryDeployments 批量重试部署
func (s *DeploymentImpl) BatchRetryDeployments(modelIDs []string) error {
	for _, modelID := range modelIDs {
		// 这里需要获取模型的DDL语句并重试部署
		// 简化实现，跳过具体逻辑
		fmt.Printf("重试部署模型: %s\n", modelID)
	}
	return nil
}

// getExistingColumns 查询表的现有列
func (s *DeploymentImpl) getExistingColumns(
	ctx context.Context,
	projectID, databaseName, tableName string,
) (map[string]bool, error) {
	logger := logfacade.GetLogger(ctx)

	// Get orgName from context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取组织名称失败: %v", err)
	}

	// 获取数据库连接
	conn, err := s.clusterManager.GetConnectionWithDatabase(ctx, orgName, projectID, databaseName)
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %v", err)
	}

	// 查询表的列信息
	query := `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
	          WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`

	rows, err := conn.QueryContext(ctx, query, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询表列信息失败: %v", err)
	}
	defer rows.Close()

	// 构建列名集合
	existingColumns := make(map[string]bool)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("扫描列名失败: %v", err)
		}
		existingColumns[columnName] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果失败: %v", err)
	}

	logger.Infof(ctx, "表 %s.%s 现有列: %v", databaseName, tableName, existingColumns)
	return existingColumns, nil
}

// DeployModelToAddFields 部署模型以添加字段
func (s *DeploymentImpl) DeployModelToAddFields(
	ctx context.Context,
	model *modeldesign.DataModel,
	addFields []*modeldesign.FieldDefinition,
) error {
	logger := logfacade.GetLogger(ctx)

	// 查询表的现有列
	existingColumns, err := s.getExistingColumns(
		ctx,
		model.ProjectSlug,
		model.DatabaseName,
		model.ModelName,
	)
	if err != nil {
		return err
	}

	// 过滤掉已存在的列和关系字段
	fieldsToAdd := make([]*modeldesign.FieldDefinition, 0, len(addFields))
	for _, field := range addFields {
		// 跳过关系字段 - 不需要部署到客户DB
		if field.IsRelationField() {
			logger.Infof(ctx, "跳过关系字段 %s", field.Name)
			continue
		}
		if existingColumns[field.Name] {
			logger.Infof(ctx, "列 %s 已存在，跳过添加", field.Name)
			continue
		}
		fieldsToAdd = append(fieldsToAdd, field)
	}

	// 如果没有需要添加的列，直接返回
	if len(fieldsToAdd) == 0 {
		logger.Infof(ctx, "所有列都已存在，无需添加")
		return nil
	}

	logger.Infof(ctx, "需要添加的列: %d 个", len(fieldsToAdd))

	// 生成DDL语句
	converter := newDdlConverter(ctx)
	ddl, err := converter.GenerateAddColumns(model, fieldsToAdd)
	if err != nil {
		return shared.NewRepositoryError(
			shared.ErrTypeSQLConvertion,
			fmt.Sprintf("生成添加列DDL失败: %v", err),
		)
	}

	// 执行添加列的DDL语句
	repositoryErr := s.deployToClientDatabase(
		ctx,
		model.GetModelLocator(),
		ddl,
		modeldesign.ChangeAddField,
	)
	if repositoryErr != nil {
		return repositoryErr
	}
	return nil
}
