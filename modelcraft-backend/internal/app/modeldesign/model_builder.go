package modeldesign

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// validateModelDatabaseName 校验数据库表名格式
// 允许以字母或下划线开头，只能包含字母、数字和下划线
func validateModelDatabaseName(name string) error {
	if name == "" {
		return bizerrors.NewValidationError("模型数据库名称不能为空")
	}

	matched, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if err != nil {
		return fmt.Errorf("校验模型数据库名称时发生错误: %v", err)
	}

	if !matched {
		return bizerrors.NewValidationError("模型数据库名称格式不正确：必须以字母或下划线开头，只能包含字母、数字和下划线")
	}

	return nil
}

// validateModelDisplayName 校验模型显示名称格式
// 必须以字母开头，只能包含字母、数字和下划线（不允许下划线开头）
func validateModelDisplayName(name string) error {
	if name == "" {
		return bizerrors.NewValidationError("模型名称不能为空")
	}

	matched, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, name)
	if err != nil {
		return fmt.Errorf("校验模型名称时发生错误: %v", err)
	}

	if !matched {
		return bizerrors.NewValidationError("模型名称格式不正确：必须以字母开头，只能包含字母、数字和下划线")
	}

	return nil
}

func newModelFromCommand(ctx context.Context, cmd CreateModelCommand) (*modeldesign.DataModel, error) {
	now := time.Now()
	if err := validateModelDisplayName(cmd.Name); err != nil {
		return nil, err
	}
	if err := validateModelDatabaseName(cmd.DatabaseName); err != nil {
		return nil, err
	}

	// 使用工厂方法创建 ModelLocator
	locator, err := modeldesign.NewModelLocator(cmd.OrgName, cmd.ProjectSlug, cmd.DatabaseName, cmd.Name)
	if err != nil {
		return nil, err
	}

	m := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID:               uuid.New().String(),
			ModelLocator:     *locator,
			Title:            cmd.Title,
			Description:      cmd.Description,
			StorageType:      cmd.StorageType,
			DisplayField:     cmd.DisplayField,
			Version:          1, // 初始版本号为1
			Status:           "draft",
			DeploymentStatus: modeldesign.DeploymentPending, // 新创建的模型待同步
			CreatedVia:       modeldesign.ModelCreationSourceNew,
			LastSyncAt:       nil,
			SyncError:        "",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		Fields: make([]*modeldesign.FieldDefinition, 0),
	}

	return m, nil
}

// SkippedFieldInfo 跳过的字段信息
type SkippedFieldInfo struct {
	FieldName string
	MySQLType string
	Reason    string
}

// BuildModelFromTableResult 从表构建模型的结果
type BuildModelFromTableResult struct {
	Model         *modeldesign.DataModel
	SkippedFields []SkippedFieldInfo
}

// BuildModelFromTable 从表定义构建模型
func BuildModelFromTable(
	tableName string,
	tableComment string,
	columns []modeldesign.TableColumn,
	primaryKeys []string,
	orgName string,
	projectSlug string,
	databaseName string,
) (*BuildModelFromTableResult, error) {
	now := time.Now()

	// 生成模型元数据
	modelName := normalizeModelName(tableName)
	modelTitle := formatModelTitle(tableName)
	modelDescription := tableComment
	if modelDescription == "" {
		modelDescription = fmt.Sprintf("Model reverse-engineered from table %s", tableName)
	}

	// 创建 ModelLocator
	locator, err := modeldesign.NewModelLocator(orgName, projectSlug, databaseName, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to NewModelLocator: %w", err)
	}

	// 创建模型ID
	modelID := uuid.New().String()

	// 创建 DataModel
	model := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID:               modelID,
			ModelLocator:     *locator,
			Title:            modelTitle,
			Description:      modelDescription,
			StorageType:      "mysql",
			Version:          1,
			Status:           "draft",
			DeploymentStatus: modeldesign.DeploymentPending,
			LastSyncAt:       nil,
			SyncError:        "",
			CreatedAt:        now,
			UpdatedAt:        now,
			CreatedVia:       modeldesign.ModelCreationSourceImported,
		},
		Fields: make([]*modeldesign.FieldDefinition, 0),
	}

	// 创建反向类型映射器
	reverseMapper := modeldesign.NewReverseTypeMapper()

	// 转换列为字段
	skippedFields := []SkippedFieldInfo{}
	primaryKeyMap := make(map[string]bool)
	for _, pk := range primaryKeys {
		primaryKeyMap[pk] = true
	}

	// 转换每个列
	for _, col := range columns {
		// 构建字段定义
		fieldDef, mappingResult := reverseMapper.BuildFieldDefinition(col, modelID, locator)

		if !mappingResult.Supported {
			// 记录跳过的字段
			skippedFields = append(skippedFields, SkippedFieldInfo{
				FieldName: col.Name,
				MySQLType: col.DataType,
				Reason:    mappingResult.SkipReason,
			})
			continue
		}

		// 导入场景需要保留主键信息，但不注入系统字段。
		if primaryKeyMap[col.Name] {
			fieldDef.IsPrimary = true
			fieldDef.IsUnique = true
			fieldDef.NonNull = true
		}

		// 设置时间戳
		fieldDef.CreatedAt = now
		fieldDef.UpdatedAt = now

		model.Fields = append(model.Fields, fieldDef)
	}

	// 如果所有字段都被跳过，返回错误
	if len(model.Fields) == 0 {
		return nil, fmt.Errorf("all fields were skipped, cannot create model")
	}

	return &BuildModelFromTableResult{
		Model:         model,
		SkippedFields: skippedFields,
	}, nil
}

// normalizeModelName 规范化模型名称（转为小写）
func normalizeModelName(tableName string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(tableName, "_")
}

// formatModelTitle 格式化模型标题
// 例如: "user_info" -> "User Info"
func formatModelTitle(tableName string) string {
	// 移除前缀（如 tbl_, t_）
	name := regexp.MustCompile(`^(tbl_|t_)`).ReplaceAllString(tableName, "")

	// 按下划线分割
	parts := regexp.MustCompile(`[_\s]+`).Split(name, -1)

	// 首字母大写
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = string(part[0]-32) + part[1:] // 简单首字母大写
			if part[0] >= 'a' && part[0] <= 'z' {
				parts[i] = string(part[0]-32) + part[1:]
			} else {
				parts[i] = part
			}
		}
	}

	return regexp.MustCompile(`\s+`).ReplaceAllString(fmt.Sprintf("%s", parts), " ")
}

// findColumn 查找列
func findColumn(columns []modeldesign.TableColumn, name string) *modeldesign.TableColumn {
	for _, col := range columns {
		if col.Name == name {
			return &col
		}
	}
	return nil
}
