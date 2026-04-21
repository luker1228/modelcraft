package modeldesign

import (
	"context"
	"errors"
	"fmt"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/database/ddl"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// ReverseEngineerAppService 反向工程应用服务
type ReverseEngineerAppService struct {
	modelAppService    *ModelDesignAppService
	schemaIntrospector ddl.SchemaIntrospector
	clusterManager     *repository.ClusterConnectionManager
	clusterRepo        cluster.DatabaseClusterRepository
	modelRepo          modeldesign.ModelRepository
}

// NewReverseEngineerAppService 创建反向工程应用服务实例
func NewReverseEngineerAppService(
	modelAppService *ModelDesignAppService,
	clusterManager *repository.ClusterConnectionManager,
	clusterRepo cluster.DatabaseClusterRepository,
	modelRepo modeldesign.ModelRepository,
) *ReverseEngineerAppService {
	return &ReverseEngineerAppService{
		modelAppService:    modelAppService,
		schemaIntrospector: ddl.NewSchemaIntrospector(),
		clusterManager:     clusterManager,
		clusterRepo:        clusterRepo,
		modelRepo:          modelRepo,
	}
}

// ImportModelCommand 导入模型命令
type ImportModelCommand struct {
	OrgName      string
	ProjectSlug  string
	DatabaseName string
	TableName    string
}

// ImportModelResult 导入模型结果
type ImportModelResult struct {
	ModelID       string
	ModelName     string
	FieldsCount   int
	SkippedFields []SkippedFieldInfo
}

// ImportModel 导入模型
func (s *ReverseEngineerAppService) ImportModel(
	ctx context.Context,
	cmd ImportModelCommand,
) (*ImportModelResult, error) {
	logger := logfacade.GetLogger(ctx)

	// 1. 验证请求参数
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}

	// 2. 获取表定义

	tableDef, err := s.getTableDefinition(ctx, cmd)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Retrieved table definition: %s with %d columns", tableDef.TableName, len(tableDef.Columns))

	// 3. 检查模型名冲突
	if err := s.checkModelNameConflict(
		ctx,
		cmd.OrgName,
		tableDef.TableName,
		cmd.DatabaseName,
		cmd.ProjectSlug,
	); err != nil {
		return nil, err
	}

	// 4. 从表定义构建模型
	buildResult, err := s.buildModelFromTable(tableDef, cmd)
	if err != nil {
		return nil, err
	}

	logger.Infof(
		ctx, "Built model with %d fields, %d skipped",
		len(buildResult.Model.Fields),
		len(buildResult.SkippedFields),
	)

	// 5. 创建模型（复用现有的创建逻辑）
	createdModel, err := s.createModel(ctx, cmd.OrgName, buildResult.Model)
	if err != nil {
		return nil, err
	}

	// 6. 返回结果
	return &ImportModelResult{
		ModelID:       createdModel.ID,
		ModelName:     createdModel.ModelLocator.ModelName,
		FieldsCount:   len(createdModel.Fields),
		SkippedFields: buildResult.SkippedFields,
	}, nil
}

// validateCommand 验证命令参数
func (s *ReverseEngineerAppService) validateCommand(cmd ImportModelCommand) error {
	if cmd.DatabaseName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "database_name is required")
	}
	if cmd.TableName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "table_name is required")
	}
	return nil
}

// getTableDefinition 从数据库内省获取表定义
func (s *ReverseEngineerAppService) getTableDefinition(
	ctx context.Context,
	cmd ImportModelCommand,
) (*ddl.TableDefinition, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "Introspecting table: %s", cmd.TableName)

	sqlDB, err := s.clusterManager.GetConnectionWithDatabase(
		ctx,
		cmd.OrgName,
		cmd.ProjectSlug,
		cmd.DatabaseName,
	)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get database connection for %s", cmd.DatabaseName)
	}

	tableDef, err := s.schemaIntrospector.IntrospectTable(ctx, sqlDB, cmd.TableName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to introspect table: %s", cmd.TableName)
	}

	return tableDef, nil
}

// checkModelNameConflict 检查模型名冲突
func (s *ReverseEngineerAppService) checkModelNameConflict(
	ctx context.Context,
	orgName string,
	tableName string,
	databaseName string,
	projectId string,
) error {
	// 规范化模型名
	modelName := normalizeModelName(tableName)

	// 查询是否已存在同名模型
	opts := modeldesign.NewModelQueryOptions()
	existingModel, err := s.modelRepo.GetByName(ctx, orgName, databaseName, modelName, projectId, opts)
	if err != nil {
		// 未找到属于正常情况，说明没有冲突，可以继续创建
		if shared.IsNotFoundError(err) {
			return nil
		}
		return err
	}

	if existingModel != nil {
		return bizerrors.NewError(bizerrors.ModelAlreadyExists, fmt.Sprintf("model %s already exists", modelName))
	}

	return nil
}

// buildModelFromTable 从表定义构建模型
func (s *ReverseEngineerAppService) buildModelFromTable(
	tableDef *ddl.TableDefinition,
	cmd ImportModelCommand,
) (*BuildModelFromTableResult, error) {
	// 转换 ddl.ColumnDefinition 到 modeldesign.TableColumn
	columns := make([]modeldesign.TableColumn, 0, len(tableDef.Columns))
	for _, col := range tableDef.Columns {
		columns = append(columns, modeldesign.TableColumn{
			Name:          col.Name,
			DataType:      col.DataType,
			Length:        col.Length,
			Precision:     col.Precision,
			Scale:         col.Scale,
			Nullable:      col.Nullable,
			DefaultValue:  col.DefaultValue,
			AutoIncrement: col.AutoIncrement,
			Comment:       col.Comment,
		})
	}

	// 调用 BuildModelFromTable
	return BuildModelFromTable(
		tableDef.TableName,
		tableDef.Comment,
		columns,
		tableDef.PrimaryKeys,
		cmd.OrgName,
		cmd.ProjectSlug,
		cmd.DatabaseName,
	)
}

// ImportPrivateModel imports a table from a private database as a model.
// Implements project.PrivateModelImporter. Skips if the model already exists.
func (s *ReverseEngineerAppService) ImportPrivateModel(
	ctx context.Context,
	orgName, projectSlug, databaseName, tableName string,
) error {
	_, err := s.ImportModel(ctx, ImportModelCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
		TableName:    tableName,
	})
	if err != nil {
		// Model already exists is not an error in this context.
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) && bizErr.Info().GetCode() == bizerrors.ModelAlreadyExists.GetCode() {
			return nil
		}
		return err
	}
	return nil
}
func (s *ReverseEngineerAppService) createModel(
	ctx context.Context,
	orgName string,
	model *modeldesign.DataModel,
) (*modeldesign.DataModel, error) {
	// 验证模型
	if err := model.Validate(); err != nil {
		return nil, err
	}

	// 在事务中创建模型并部署
	if err := s.modelAppService.transactionDeployModel(ctx, orgName, model); err != nil {
		return nil, err
	}

	// 获取创建后的完整模型
	opt := modeldesign.NewModelQueryOptions().WithFields()
	return s.modelRepo.GetByID(ctx, model.ID, opt)
}
