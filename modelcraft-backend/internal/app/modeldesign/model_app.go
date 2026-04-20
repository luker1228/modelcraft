package modeldesign

import (
	"context"
	"errors"
	"fmt"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/lexorder"
	"modelcraft/pkg/logfacade"
	"strings"

	rlsdomain "modelcraft/internal/domain/rls"

	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// ModelDesignAppService 模型设计应用服务，负责模型在平台数据库和客户数据库之间的同步操作
type ModelDesignAppService struct {
	deployRepo    modeldesign.DeployRepo
	modelRepo     modeldesign.ModelRepository
	fkRepo        modeldesign.LogicalForeignKeyRepository
	clusterRepo   cluster.DatabaseClusterRepository
	txManager     repository.TxManager
	enumAssocRepo modeldesign.FieldEnumAssociationRepository
	enumRepo      modeldesign.EnumRepository
}

// ModelDesignAppServiceDeps defines constructor dependencies.
type ModelDesignAppServiceDeps struct {
	DeployRepo    modeldesign.DeployRepo
	ModelRepo     modeldesign.ModelRepository
	FKRepo        modeldesign.LogicalForeignKeyRepository
	ClusterRepo   cluster.DatabaseClusterRepository
	TxManager     repository.TxManager
	EnumAssocRepo modeldesign.FieldEnumAssociationRepository
	EnumRepo      modeldesign.EnumRepository
}

// NewModelDesignAppService 创建模型设计应用服务实例
func NewModelDesignAppService(deps ModelDesignAppServiceDeps) *ModelDesignAppService {
	return &ModelDesignAppService{
		txManager:     deps.TxManager,
		deployRepo:    deps.DeployRepo,
		modelRepo:     deps.ModelRepo,
		fkRepo:        deps.FKRepo,
		clusterRepo:   deps.ClusterRepo,
		enumAssocRepo: deps.EnumAssocRepo,
		enumRepo:      deps.EnumRepo,
	}
}

// getLogger 获取logger，优先使用context中的logger，降级到baseLogger
func (s *ModelDesignAppService) getLogger(ctx context.Context) logfacade.Logger {
	return logfacade.GetLogger(ctx)
}

// CreateModelSync 同步创建模型（平台DB+客户DB）
func (s *ModelDesignAppService) CreateModelSync(
	ctx context.Context,
	cmd CreateModelCommand,
) (string, error) {
	// Get orgName from context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get organization name: %w", err)
	}
	_, err = s.clusterRepo.GetByProjectKey(ctx, orgName, cmd.ProjectSlug)
	if err != nil {
		return "", bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound, cmd.ProjectSlug)
	}
	model, err := newModelFromCommand(ctx, cmd)
	if err != nil {
		return "", err
	}
	fields := modeldesign.GetNewModelSystemFields()
	model.AddFields(fields)
	s.getLogger(ctx).Infof(ctx, "model: %s", bizutils.MarshalToStringIgnoreErr(model))
	err = model.Validate()
	if err != nil {
		return "", bizerrors.Wrapf(err, "模型规则校验失败")
	}

	// 验证 displayField 有效性（必须是模型字段中存在且可字符串化的字段）
	if err := model.ValidateDisplayField(); err != nil {
		return "", err
	}

	// 1. Check model name uniqueness
	existingModel, err := s.modelRepo.GetByName(ctx, orgName, model.DatabaseName, model.ModelName, model.ProjectSlug)
	if err != nil && !shared.IsNotFoundError(err) {
		return "", fmt.Errorf("failed to check model name uniqueness: %w", err)
	}
	if err == nil && existingModel != nil {
		return "", bizerrors.NewErrorFromContext(ctx, bizerrors.ModelAlreadyExists, model.GetBizUniqueName())
	}

	// 2. Check table existence in client DB
	tableExists, err := s.deployRepo.CheckTableExists(ctx, model)
	if err != nil {
		return "", fmt.Errorf("failed to check table existence: %w", err)
	}
	if tableExists {
		return "", bizerrors.NewErrorFromContext(ctx, bizerrors.ModelTableAlreadyExists,
			model.ModelName, model.DatabaseName)
	}

	// 3. Deploy in transaction
	if err = s.transactionDeployModel(ctx, orgName, model); err != nil {
		return "", err
	}

	return model.ID, nil
}

// transactionDeployModel 在事务中保存模型并部署到客户DB
func (s *ModelDesignAppService) transactionDeployModel(
	ctx context.Context,
	orgName string,
	model *modeldesign.DataModel,
) error {
	hasOwnerField := model.IsRLSEnabled()

	return s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		modelRepository := repository.NewSqlModelDesignRepository(q)
		if err := modelRepository.Save(ctx, orgName, model); err != nil {
			if shared.IsRepoError(err, shared.ErrTypeConstraint) {
				return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelAlreadyExists, model.GetBizUniqueName())
			}
			return err
		}

		opt := modeldesign.NewModelQueryOptions().WithFields()
		createdModel, err := modelRepository.GetByID(ctx, model.ID, opt)
		if err != nil {
			return err
		}
		if createdModel == nil {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, model.ID)
		}

		if hasOwnerField {
			policy := &modeldesign.ModelRLSPolicy{ModelID: model.ID}
			policy.ApplyPreset(rlsdomain.RLSPresetReadWriteOwner)
			if err := q.UpsertModelRLSPolicy(ctx, dbgen.UpsertModelRLSPolicyParams{
				ModelID:         model.ID,
				SelectPredicate: string(policy.SelectPredicate),
				InsertCheck:     string(policy.InsertCheck),
				UpdatePredicate: string(policy.UpdatePredicate),
				UpdateCheck:     string(policy.UpdateCheck),
				DeletePredicate: string(policy.DeletePredicate),
			}); err != nil {
				return fmt.Errorf("failed to upsert default rls policy: %w", err)
			}
		}

		return s.deployRepo.DeployModelToCreate(ctx, createdModel)
	})
}

// UpdateModelMeta 更新模型元数据（仅平台DB）
func (s *ModelDesignAppService) UpdateModelMeta(ctx context.Context, id string, cmd UpdateModelMetaCommand) error {
	// 获取模型（需要字段信息以验证 displayField）
	opts := modeldesign.NewModelQueryOptions().WithFields()
	model, repoErr := s.modelRepo.GetByID(ctx, id, opts)
	if repoErr != nil {
		return repoErr
	}
	if model == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, id)
	}

	// 更新模型元信息（title/description）
	if err := model.Update(cmd.Title, cmd.Description); err != nil {
		return fmt.Errorf("更新模型属性失败: %w", err)
	}

	// 更新 displayField（如果传入非 nil 则更新）
	if cmd.DisplayField != nil {
		model.UpdateDisplayField(cmd.DisplayField)
	}

	// 验证元数据
	if err := model.ValidateMeta(); err != nil {
		return err
	}

	// 验证 displayField 有效性
	if err := model.ValidateDisplayField(); err != nil {
		return err
	}

	// 保存到数据库
	if repoErr := s.modelRepo.Update(ctx, model); repoErr != nil {
		return repoErr
	}
	return nil
}

// DeleteModelSync 同步删除模型。
// 当 dropTable=true 时同时执行 DROP TABLE DDL；dropTable=false（默认）时只删除元数据，不删除底层表。
func (s *ModelDesignAppService) DeleteModelSync(ctx context.Context, id, projectSlug string, dropTable bool) error {
	logger := logfacade.GetLogger(ctx)
	model, err := s.modelRepo.GetByID(ctx, id)
	if err != nil {
		logger.Infof(ctx, "err = %+v, %w, %T", err, err, err)
		return err
	}
	if model == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, id)
	}
	if dropTable {
		err = s.deployRepo.DeployModelToDrop(ctx, model)
		if err != nil {
			return err
		}
	}
	if err := s.modelRepo.Delete(ctx, id); err != nil {
		return err
	}

	// 4.10: 清理孤立的反向 FK 行（DB CASCADE 已删除本模型的 FK 行，
	// 但 ref_model_id = deleted_model_id 的反向行可能残留）
	if s.fkRepo != nil {
		s.cleanOrphanedFKRows(ctx, model.OrgName, id)
	}
	return nil
}

// cleanOrphanedFKRows removes orphaned FK rows left after model deletion.
// DB CASCADE removes FK rows where model_id = deleted model, but rows where
// ref_model_id = deleted model may remain as orphans.
func (s *ModelDesignAppService) cleanOrphanedFKRows(ctx context.Context, orgName, modelID string) {
	logger := logfacade.GetLogger(ctx)
	orphans, err := s.fkRepo.FindByModel(ctx, orgName, modelID)
	if err != nil {
		logger.Warnf(ctx, "DeleteModelSync: failed to query orphaned FK rows for model %s: %v", modelID, err)
		return
	}
	seen := make(map[string]bool)
	for _, row := range orphans {
		if seen[row.PairID] {
			continue
		}
		seen[row.PairID] = true
		if delErr := s.fkRepo.DeleteByPairID(ctx, orgName, row.PairID); delErr != nil {
			logger.Warnf(ctx, "DeleteModelSync: failed to delete orphaned FK pair %s: %v", row.PairID, delErr)
		}
	}
}

// QueryModels 查询模型列表
func (s *ModelDesignAppService) QueryModels(
	ctx context.Context,
	domainQuery modeldesign.ModelQuery,
) ([]modeldesign.DataModel, int, error) {
	// Validate and set defaults for query
	if err := domainQuery.Validate(); err != nil {
		return nil, 0, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, fmt.Sprintf("查询参数验证失败: %v", err))
	}
	domainQuery.SetDefaults()

	// 查询模型列表
	modelQueryResults, total, repoErr := s.modelRepo.Query(ctx, domainQuery)
	if repoErr != nil {
		return nil, 0, repoErr
	}
	return modelQueryResults, total, nil
}

// QueryModelsWithCommand 查询模型列表（从 Command 转换）
func (s *ModelDesignAppService) QueryModelsWithCommand(
	ctx context.Context,
	cmd ModelQueryCommand,
) ([]modeldesign.DataModel, int, error) {
	// Get OrgName from context if not provided in command
	orgName := cmd.OrgName
	if orgName == "" {
		var err error
		orgName, err = ctxutils.GetOrgNameFromContext(ctx)
		if err != nil {
			return nil, 0, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
		}
	}

	// 将Command转换为Domain查询对象
	domainQuery := modeldesign.ModelQuery{
		OrgName:      orgName,
		ProjectSlug:  cmd.ProjectSlug,
		DatabaseName: cmd.DatabaseName,
		Name:         cmd.Name,
		Title:        cmd.Title,
		Status:       cmd.Status,
		StorageType:  cmd.StorageType,
		Page:         cmd.Page,
		PageSize:     cmd.PageSize,
	}
	domainQuery.SetDefaults()

	if err := domainQuery.Validate(); err != nil {
		return nil, 0, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, fmt.Sprintf("查询参数验证失败: %v", err))
	}

	return s.QueryModels(ctx, domainQuery)
}

// QueryDatabaseCatalogWithCommand 查询项目下可用数据库目录（分页）。
func (s *ModelDesignAppService) QueryDatabaseCatalogWithCommand(
	ctx context.Context,
	cmd DatabaseCatalogQueryCommand,
) ([]string, int, error) {
	if cmd.OrgName == "" {
		return nil, 0, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "orgName is required")
	}
	if cmd.ProjectSlug == "" {
		return nil, 0, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "projectSlug is required")
	}

	page := cmd.Page
	if page <= 0 {
		page = 1
	}
	pageSize := cmd.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return s.modelRepo.ListDatabaseCatalog(ctx, cmd.OrgName, cmd.ProjectSlug, cmd.Search, page, pageSize)
}

// AddFieldsWithResults 按字段独立处理添加请求，并返回逐字段结果。
func (s *ModelDesignAppService) AddFieldsWithResults(
	ctx context.Context,
	cmd AddFieldCommand,
) ([]*AddFieldItemResult, error) {
	results := make([]*AddFieldItemResult, 0, len(cmd.Fields))
	for _, field := range cmd.Fields {
		if field == nil {
			results = append(results, &AddFieldItemResult{
				Name:    "",
				Success: false,
				Err:     bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "field input is required"),
			})
			continue
		}

		singleCmd := AddFieldCommand{
			ModelID: cmd.ModelID,
			Fields:  []*modeldesign.FieldDefinition{field},
		}
		err := s.AddFieldSync(ctx, singleCmd)
		results = append(results, &AddFieldItemResult{
			Name:    field.Name,
			Success: err == nil,
			Err:     err,
		})
	}
	return results, nil
}

// AddFieldSync 同步添加字段（支持物理字段和关联字段）
func (s *ModelDesignAppService) AddFieldSync(ctx context.Context, cmd AddFieldCommand) error {
	logger := logfacade.GetLogger(ctx)
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization name: %w", err)
	}
	modelID := cmd.ModelID

	opts := modeldesign.NewModelQueryOptions().WithFields()
	model, err := s.modelRepo.GetByID(ctx, modelID, opts)
	if err != nil {
		return err
	}
	if model == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
	}

	toAddFields := cmd.Fields
	// 从 model 补全字段的 ModelID 和 ModelLocator，resolver 层无法提前得知这些信息
	for _, field := range toAddFields {
		field.ModelID = model.ID
		field.ModelLocator = model.GetModelLocator()
	}

	// 过滤掉与系统字段同名的字段（静默跳过，不报错）
	systemFieldNames := lo.SliceToMap(
		modeldesign.GetSystemFields(),
		func(f *modeldesign.FieldDefinition) (string, struct{}) {
			return f.Name, struct{}{}
		},
	)
	toAddFields = lo.Filter(toAddFields, func(f *modeldesign.FieldDefinition, _ int) bool {
		_, isSystem := systemFieldNames[f.Name]
		return !isSystem
	})
	if len(toAddFields) == 0 {
		return nil
	}

	fieldAbstracts := lo.SliceToMap(
		toAddFields,
		func(field *modeldesign.FieldDefinition) (string, *modeldesign.FieldType) {
			return field.Name, field.Type
		},
	)
	logger.Infof(ctx, "add fields: %s", bizutils.MarshalToStringIgnoreErr(fieldAbstracts))

	// 交叉验证：检查要添加的字段是否与现有字段重复
	fieldService := modeldesign.NewFieldService()
	if err := fieldService.ValidateAddFieldsNotExist(
		model.Fields,
		fieldService.GetNamesFromFields(toAddFields)...,
	); err != nil {
		return err
	}

	var physicFields []*modeldesign.FieldDefinition
	var relationFields []*modeldesign.FieldDefinition
	for _, field := range toAddFields {
		switch {
		case field.IsRelationField():
			relationFields = append(relationFields, field)
		default:
			physicFields = append(physicFields, field)
		}
	}
	// 添加本表字段
	err = s.addPhysicFields(ctx, orgName, model, physicFields)
	if err != nil {
		return err
	}
	// 添加 RELATION 格式字段（新方式：通过 relate_fk_id 关联逻辑外键）
	err = s.addRelationFields(ctx, orgName, model, relationFields)
	if err != nil {
		return err
	}

	return nil
}

// addPhysicFields 添加物理字段
func (s *ModelDesignAppService) addPhysicFields(
	ctx context.Context,
	orgName string,
	model *modeldesign.DataModel,
	fieldEntitys []*modeldesign.FieldDefinition,
) error {
	if len(fieldEntitys) == 0 {
		return nil
	}
	log := logfacade.GetLogger(ctx)
	log.Infof(ctx, "add physic fields: %s", bizutils.MarshalToStringIgnoreErr(fieldEntitys))

	// Infer IsArray from EnumDefinition.IsMultiSelect for ENUM fields
	if err := s.inferEnumFieldIsArray(ctx, model, fieldEntitys); err != nil {
		return err
	}

	for _, field := range fieldEntitys {
		if err := field.Validate(); err != nil {
			return err
		}
	}
	if err := s.assignDisplayOrders(ctx, model.ID, fieldEntitys); err != nil {
		return err
	}
	if err := s.modelRepo.AddFields(ctx, orgName, fieldEntitys); err != nil {
		return err
	}

	// 部署到客户 DB，若失败则回滚已添加的字段（补偿机制）
	fieldService := modeldesign.NewFieldService()
	fieldNames := fieldService.GetNamesFromFields(fieldEntitys)

	err := s.deployRepo.DeployModelToAddFields(ctx, model, fieldEntitys)
	if err != nil {
		// 补偿：删除已落库但部署失败的字段，避免状态污染
		log.Warnf(ctx, "部署字段失败，执行回滚: modelID=%s, fields=%v, err=%v", model.ID, fieldNames, err)
		rollbackErr := s.modelRepo.DeleteFields(ctx, model.ID, fieldNames)
		if rollbackErr != nil {
			log.Errorf(ctx, "回滚字段失败: modelID=%s, fields=%v, err=%v", model.ID, fieldNames, rollbackErr)
			// 返回组合错误：包含部署失败和回滚失败
			return bizerrors.Wrapf(errors.Join(err, rollbackErr),
				"部署模型到客户DB失败且回滚失败: modelID=%s, fields=%v", model.ID, fieldNames)
		}
		return fmt.Errorf("部署模型到客户DB失败: %w", err)
	}

	updateReq := modeldesign.UpdateFieldsStatusRequest{
		ModelId: model.ID,
		Name:    fieldNames,
		Status:  modeldesign.FieldStatusToDelete,
	}
	if err := s.modelRepo.UpdateFieldsStatus(ctx, updateReq); err != nil {
		return fmt.Errorf("更新字段状态失败: %w", err)
	}

	// 创建枚举字段的关联记录
	for _, field := range fieldEntitys {
		if field.IsEnumField() && field.EnumName != "" {
			enumAssocRepo := s.enumAssocRepo
			association := &modeldesign.FieldEnumAssociation{
				ModelID:   model.ID,
				FieldName: field.Name,
				ProjectScope: project.ProjectScope{
					OrgName:     orgName,
					ProjectSlug: model.ProjectSlug,
				},
				EnumName:     field.EnumName,
				DatabaseName: model.DatabaseName,
			}
			if err := enumAssocRepo.Create(ctx, association); err != nil {
				log.Errorf(ctx, "创建枚举关联失败: modelID=%s, fieldName=%s, enumName=%s, err=%v",
					model.ID, field.Name, field.EnumName, err)
				return fmt.Errorf("创建枚举关联失败: %w", err)
			}
			log.Infof(ctx, "创建枚举关联成功: modelID=%s, fieldName=%s, enumName=%s",
				model.ID, field.Name, field.EnumName)
		}
	}

	return nil
}

// inferEnumFieldIsArray infers and validates the IsArray field for ENUM format fields
// based on the associated EnumDefinition.IsMultiSelect property.
//
// Rules:
// 1. For ENUM format fields with an associated enum (EnumName != ""):
//   - If enum.IsMultiSelect is true, field.IsArray is set to true
//   - If enum.IsMultiSelect is false, field.IsArray must be false (error if true)
//
// 2. Non-ENUM fields are not affected by this logic.
// 3. ENUM fields without EnumName are skipped (legacy inline enum values).
func (s *ModelDesignAppService) inferEnumFieldIsArray(
	ctx context.Context,
	model *modeldesign.DataModel,
	fields []*modeldesign.FieldDefinition,
) error {
	// Skip if enumRepo is not configured
	if s.enumRepo == nil {
		return nil
	}

	for _, field := range fields {
		// Only process ENUM format fields with an associated enum
		if !field.IsEnumField() || field.EnumName == "" {
			continue
		}

		// Fetch the enum definition
		enumDef, err := s.enumRepo.FindByName(model.OrgName, model.ProjectSlug, field.EnumName)
		if err != nil {
			return bizerrors.Wrapf(err, "failed to find enum '%s' for field '%s'", field.EnumName, field.Name)
		}
		if enumDef == nil {
			return bizerrors.NewError(bizerrors.EnumNotFound, field.EnumName)
		}

		// Validate and infer IsArray based on enum.IsMultiSelect
		if field.IsArray && !enumDef.IsMultiSelect {
			// User requested IsArray=true, but enum does not support multi-select
			return bizerrors.Errorf(
				"field '%s': enum '%s' does not support multi-select, cannot set isArray=true",
				field.Name, field.EnumName,
			)
		}

		// Infer IsArray from enum.IsMultiSelect
		field.IsArray = enumDef.IsMultiSelect
	}

	return nil
}

// addRelationFields 添加 RELATION 格式字段（新方式：通过 relate_fk_id 关联逻辑外键）
// RELATION 格式字段只保存到平台 DB，不部署到客户 DB（它是逻辑字段，不对应物理列）
func (s *ModelDesignAppService) addRelationFields(
	ctx context.Context,
	orgName string,
	model *modeldesign.DataModel,
	toAddFields []*modeldesign.FieldDefinition,
) error {
	if len(toAddFields) == 0 {
		return nil
	}
	log := logfacade.GetLogger(ctx)

	// Build FK validator if fkRepo is available
	var fkSvc *AddFieldFKService
	if s.fkRepo != nil {
		fkSvc = newAddFieldFKService(s.modelRepo, s.fkRepo)
	}

	fieldService := modeldesign.NewFieldService()
	for _, field := range toAddFields {
		// 验证 relate_fk_id（domain validate 已经保证 RELATION 格式必须有 relate_fk_id）
		if err := field.Validate(); err != nil {
			return err
		}
		// 额外验证：relate_fk_id 必须指向本模型的 FK 行
		if fkSvc != nil {
			if err := fkSvc.ValidateRelateFKID(ctx, orgName, model.ID, field.RelateFKID); err != nil {
				return err
			}
		}
		log.Infof(ctx, "addRelationFields: validated field %s with relate_fk_id=%v", field.Name, field.RelateFKID)
	}

	// 分配展示顺序并保存到平台 DB
	if err := s.assignDisplayOrders(ctx, model.ID, toAddFields); err != nil {
		return err
	}
	if err := s.modelRepo.AddFields(ctx, orgName, toAddFields); err != nil {
		return err
	}

	// RELATION 字段标记为已部署（不需要物理部署）
	updateReq := modeldesign.UpdateFieldsStatusRequest{
		ModelId: model.ID,
		Name:    fieldService.GetNamesFromFields(toAddFields),
		Status:  modeldesign.FieldStatusDeploySuccess,
	}
	if err := s.modelRepo.UpdateFieldsStatus(ctx, updateReq); err != nil {
		return fmt.Errorf("更新 RELATION 字段状态失败: %w", err)
	}

	log.Infof(ctx, "Successfully added %d RELATION fields to model %s", len(toAddFields), model.ModelName)
	return nil
}

// UpdateFieldSync 同步更新字段元数据（仅平台DB）
func (s *ModelDesignAppService) UpdateFieldSync(ctx context.Context, cmd UpdateFieldCommand) error {
	modelID := cmd.ModelID
	fieldName := cmd.FieldName

	// 1. 获取字段信息（现在包含 project_slug）
	field, err := s.modelRepo.GetFieldByModelID(ctx, modelID, fieldName)
	if err != nil {
		return err
	}
	if field == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.FieldNotFound)
	}

	// 2. format 与 enum 关联语义不可变
	if cmd.Format != nil && (field.Type == nil || field.Type.Format != *cmd.Format) {
		return ErrFieldFormatImmutable
	}
	if cmd.RelateEnumName != nil {
		if !field.IsEnumField() || field.EnumName != *cmd.RelateEnumName {
			return ErrFieldFormatImmutable
		}
	}

	// 3. 更新字段属性
	field.Update(cmd.Title, cmd.Description, cmd.ValidationConfig)

	// 3. 验证字段（会验证 ModelLocator，现在包含 ProjectSlug）
	if err := field.Validate(); err != nil {
		return err
	}

	// 4. 保存到数据库
	if err := s.modelRepo.UpdateField(ctx, field); err != nil {
		return err
	}
	return nil
}

// DeprecateField 将字段标记为废弃（幂等：已废弃时直接成功）
func (s *ModelDesignAppService) DeprecateField(ctx context.Context, cmd DeprecateFieldCommand) error {
	field, err := s.modelRepo.GetFieldByModelID(ctx, cmd.ModelID, cmd.FieldName)
	if err != nil {
		return err
	}
	if field == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.FieldNotFound)
	}

	field.Deprecate()

	if err := s.modelRepo.UpdateField(ctx, field); err != nil {
		return err
	}
	return nil
}

// UndeprecateField 解除字段的废弃状态（幂等：未废弃时直接成功）
func (s *ModelDesignAppService) UndeprecateField(ctx context.Context, cmd UndeprecateFieldCommand) error {
	field, err := s.modelRepo.GetFieldByModelID(ctx, cmd.ModelID, cmd.FieldName)
	if err != nil {
		return err
	}
	if field == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.FieldNotFound)
	}

	field.Undeprecate()

	if err := s.modelRepo.UpdateField(ctx, field); err != nil {
		return err
	}
	return nil
}

// RemoveFieldSync 同步删除字段（平台DB+客户DB）
func (s *ModelDesignAppService) RemoveFieldSync(
	ctx context.Context,
	cmd RemoveFieldCommand,
) error {
	logger := s.getLogger(ctx)
	modelID := cmd.ModelID
	fieldName := cmd.FieldName

	option := &modeldesign.ModelQueryOptions{
		GetFields: true,
	}
	dataModel, err := s.modelRepo.GetByID(ctx, modelID, option)
	if err != nil {
		return err
	}
	if dataModel == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound)
	}

	field := dataModel.GetField(fieldName)
	if field == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.FieldNotFound)
	}

	// 检查字段是否被 displayField 引用
	if dataModel.DisplayField != nil && *dataModel.DisplayField == fieldName {
		return bizerrors.Errorf("cannot remove field '%s': it is configured as displayField", fieldName)
	}

	// Case 1: RELATION 格式字段（relate_fk_id 字段）——直接删除，无需物理部署
	if field.IsRelationField() {
		deleteReq := modeldesign.DeleteFieldRequest{
			ModelId: modelID,
			Name:    []string{fieldName},
		}
		return s.modelRepo.BulkDeleteFields(ctx, deleteReq)
	}

	// Case 2: FK 列字段（belongs_to_fk_id 字段）——需要检查是否有 RELATION 字段引用该 FK
	if field.BelongsToFKID != nil && s.fkRepo != nil {
		if err := s.removeFKPairIfUnreferenced(ctx, dataModel.OrgName, modelID, *field.BelongsToFKID); err != nil {
			return err
		}
	}

	updateReq := modeldesign.UpdateFieldsStatusRequest{
		ModelId: modelID,
		Name:    []string{fieldName},
		Status:  modeldesign.FieldStatusToDelete,
	}
	if err := s.modelRepo.UpdateFieldsStatus(ctx, updateReq); err != nil {
		return err
	}
	if err := s.deployRepo.DeployModelToRemoveFields(
		ctx,
		dataModel,
		[]string{fieldName},
	); err != nil {
		return fmt.Errorf("部署模型到客户DB失败: %w", err)
	}

	if err := s.modelRepo.DeleteFields(ctx, modelID, []string{fieldName}); err != nil {
		return err
	}

	logger.Infof(ctx, "成功删除字段: modelID=%s, fieldName=%s", modelID, fieldName)
	return nil
}

// removeFKPairIfUnreferenced checks if the FK pair is still referenced by any RELATION fields.
// If not referenced, it deletes the FK pair. Returns an error if the FK pair is still in use.
func (s *ModelDesignAppService) removeFKPairIfUnreferenced(ctx context.Context, orgName, modelID, fkID string) error {
	relateFields, err := s.fkRepo.FindByRelateField(ctx, orgName, fkID)
	if err != nil {
		return fmt.Errorf("RemoveFieldSync: check relate fields: %w", err)
	}
	if len(relateFields) > 0 {
		return bizerrors.NewError(bizerrors.FKPairHasRelateFields, fkID)
	}
	// No RELATION field references this FK; find and delete the FK pair.
	fkRows, err := s.fkRepo.FindByModel(ctx, orgName, modelID)
	if err != nil {
		return fmt.Errorf("RemoveFieldSync: find FK rows: %w", err)
	}
	for _, row := range fkRows {
		if row.ID == fkID {
			if err := s.fkRepo.DeleteByPairID(ctx, orgName, row.PairID); err != nil {
				return fmt.Errorf("RemoveFieldSync: delete FK pair %s: %w", row.PairID, err)
			}
			break
		}
	}
	return nil
}

// SyncModelSchemaResult 同步模型Schema结果
type SyncModelSchemaResult struct {
	Model         *modeldesign.DataModel
	FieldsAdded   int
	FieldsSkipped []string
	FieldsDeleted int
	DeletedFields []string
}

// GetModelByID 根据模型ID获取模型
func (s *ModelDesignAppService) GetModelByID(
	ctx context.Context,
	id string,
	getModelOpt *GetModelOptions,
) (*modeldesign.DataModel, error) {
	// 查询模型
	var opt *modeldesign.ModelQueryOptions = modeldesign.NewModelQueryOptions()
	if getModelOpt.GetFields {
		opt = opt.WithFields()
	}
	modelQueryResult, err := s.modelRepo.GetByID(ctx, id, opt)
	if err != nil {
		return nil, err
	}
	if modelQueryResult == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, id)
	}

	// 并发填充字段的附加元数据：ENUM 定义 + 外键 x-relation（两者互相独立，无数据竞争）
	if getModelOpt.GetFields {
		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return s.fillEnumDefinitions(gCtx, modelQueryResult)
		})
		g.Go(func() error {
			return s.fillRelationMetadata(gCtx, modelQueryResult)
		})
		if err := g.Wait(); err != nil {
			return nil, err
		}
	}

	return modelQueryResult, nil
}

// fillEnumDefinitions 批量填充模型字段的 EnumDefinition，避免 N+1 查询。
func (s *ModelDesignAppService) fillEnumDefinitions(
	ctx context.Context,
	model *modeldesign.DataModel,
) error {
	if len(model.Fields) == 0 {
		return nil
	}
	// 收集需要加载的 enumName（去重）
	enumNames := make(map[string]struct{})
	for _, f := range model.Fields {
		if f.EnumName != "" {
			enumNames[f.EnumName] = struct{}{}
		}
	}
	if len(enumNames) == 0 {
		return nil
	}
	// 优先使用请求上下文的 orgName，避免模型记录存储的 org 与当前 tenant 不一致
	enumOrgName := model.OrgName
	if ctxOrg, ctxErr := ctxutils.GetOrgNameFromContext(ctx); ctxErr == nil && ctxOrg != "" {
		enumOrgName = ctxOrg
	}
	allEnums, listErr := s.enumRepo.List(enumOrgName, model.ProjectSlug)
	if listErr != nil {
		return fmt.Errorf("GetModelByID: list enums: %w", listErr)
	}
	enumMap := make(map[string]*modeldesign.EnumDefinition, len(allEnums))
	for _, e := range allEnums {
		enumMap[e.Name] = e
	}
	for _, f := range model.Fields {
		if f.EnumName != "" {
			f.Enum = enumMap[f.EnumName] // nil-safe: missing enum stays nil
		}
	}
	return nil
}

// fillRelationMetadata 为带 FK 关联信息的字段填充 x-relation 元数据。
// 查询 LogicalForeignKey 获取目标模型 ID，再查目标 DataModel 取 DatabaseName 和 ModelName，
// 写入 field.Metadata["x-relation"]，供 JSONSchemaGenerator 注入到 JSON Schema。
func (s *ModelDesignAppService) fillRelationMetadata(
	ctx context.Context,
	model *modeldesign.DataModel,
) error {
	if len(model.Fields) == 0 || s.fkRepo == nil {
		return nil
	}

	// 收集所有 BelongsToFKID（去重），批量查 LFK 再查目标模型
	type relationInfo struct {
		databaseName string
		modelName    string
		direction    string
		cardinality  string
	}
	fkIDToRelation := make(map[string]*relationInfo)

	for _, f := range model.Fields {
		var fkID string
		switch {
		case f.BelongsToFKID != nil:
			fkID = *f.BelongsToFKID
		case f.RelateFKID != nil:
			fkID = *f.RelateFKID
		default:
			continue
		}
		if _, already := fkIDToRelation[fkID]; already {
			continue
		}

		lf, err := s.fkRepo.GetByID(ctx, fkID)
		if err != nil {
			logfacade.GetLogger(ctx).Warnf(ctx,
				"fillRelationMetadata: GetByID fkID=%s err=%v", fkID, err)
			continue
		}

		refModel, err := s.modelRepo.GetByID(ctx, lf.RefModelID)
		if err != nil {
			logfacade.GetLogger(ctx).Warnf(ctx,
				"fillRelationMetadata: GetByID refModelID=%s err=%v", lf.RefModelID, err)
			continue
		}

		fkIDToRelation[fkID] = &relationInfo{
			databaseName: refModel.DatabaseName,
			modelName:    refModel.ModelName,
			direction:    string(lf.Direction),
			cardinality:  relationCardinalityFromDirection(lf.Direction),
		}
	}

	// 将 x-relation 写入字段 Metadata
	for _, f := range model.Fields {
		var fkID string
		switch {
		case f.BelongsToFKID != nil:
			fkID = *f.BelongsToFKID
		case f.RelateFKID != nil:
			fkID = *f.RelateFKID
		default:
			continue
		}
		info, ok := fkIDToRelation[fkID]
		if !ok {
			continue
		}
		if f.Metadata == nil {
			f.Metadata = make(map[string]any)
		}
		f.Metadata["x-relation"] = map[string]string{
			"databaseName": info.databaseName,
			"modelName":    info.modelName,
			"direction":    info.direction,
			"cardinality":  info.cardinality,
		}
	}
	return nil
}

func relationCardinalityFromDirection(direction modeldesign.LogicalFKDirection) string {
	if direction == modeldesign.DirectionReverse {
		return "one-to-many"
	}
	return "many-to-one"
}

// GetFieldsByModelID 根据模型ID获取所有字段
func (s *ModelDesignAppService) GetFieldsByModelID(
	ctx context.Context,
	cmd GetFieldsCommand,
) ([]*modeldesign.FieldDefinition, error) {
	// 查询模型，包含字段信息
	fields, err := s.modelRepo.GetFieldsByModelID(ctx, cmd.ModelID)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// CreateModelFromSchema 从JSON Schema创建模型
func (s *ModelDesignAppService) CreateModelFromSchema(
	ctx context.Context,
	projectSlug string,
	schemaJSON string,
	databaseName string,
) (*modeldesign.DataModel, error) {
	logger := s.getLogger(ctx)

	// 验证cluster存在
	// Get orgName from context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}
	_, err = s.clusterRepo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound, projectSlug)
	}

	// 解析schema - 模型名称从schema中提取或使用默认值
	parser := modeldesign.NewJSONSchemaParser(ctx)
	model, err := parser.ParseSchemaWithLoggerAndModelInfo(
		schemaJSON,
		logger,
		"",
		databaseName,
	)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "Failed to parse JSON Schema")
	}

	// 如果schema中没有提供模型名称，需要从其他地方获取或生成默认名称
	if model.ModelName == "" {
		// 这里可以根据需要从schema的title或其他信息生成模型名称
		// 暂时使用schema的title作为模型名称，如果title也不存在则使用默认值
		if model.Title != "" {
			model.ModelName = strings.ToLower(strings.ReplaceAll(model.Title, " ", "_"))
		} else {
			model.ModelName = "unnamed_model"
		}
	}

	// 检查模型名冲突
	existingModel, err := s.modelRepo.GetByName(ctx, orgName, databaseName, model.ModelName, projectSlug)
	if err != nil && !shared.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check model name uniqueness: %w", err)
	}
	if err == nil && existingModel != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelAlreadyExists, model.GetBizUniqueName())
	}

	// 添加系统字段
	systemFields := modeldesign.GetNewModelSystemFields()
	model.AddFields(systemFields)

	// 验证模型
	if err := model.Validate(); err != nil {
		return nil, bizerrors.Wrapf(err, "Model validation failed")
	}

	// 检查客户DB表是否已存在
	tableExists, err := s.deployRepo.CheckTableExists(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}
	if tableExists {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelTableAlreadyExists, model.ModelName, databaseName)
	}

	// 在事务中创建并部署模型
	if err = s.transactionDeployModel(ctx, orgName, model); err != nil {
		return nil, err
	}

	// 获取创建后的完整模型
	opt := modeldesign.NewModelQueryOptions().WithFields()
	createdModel, err := s.modelRepo.GetByID(ctx, model.ID, opt)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Successfully created model from schema: %s", model.GetBizUniqueName())
	return createdModel, nil
}

// SyncModelSchemaFromJSON 从JSON Schema同步模型
func (s *ModelDesignAppService) SyncModelSchemaFromJSON(
	ctx context.Context,
	modelID string,
	schemaJSON string,
	deleteExtraFields bool,
) (*SyncModelSchemaResult, error) {
	logger := s.getLogger(ctx)

	// 1. 获取现有模型（包含字段）
	opts := modeldesign.NewModelQueryOptions().WithFields()
	existingModel, err := s.modelRepo.GetByID(ctx, modelID, opts)
	if err != nil {
		return nil, err
	}
	if existingModel == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
	}

	// 2. 解析schema - 使用现有模型的集群和数据库信息
	parser := modeldesign.NewJSONSchemaParser(ctx)
	parsedModel, err := parser.ParseSchemaWithLoggerAndModelInfo(
		schemaJSON,
		logger,
		existingModel.ModelName,
		existingModel.DatabaseName,
	)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "Failed to parse JSON Schema")
	}

	// 3. 验证模型名称匹配（如果schema中提供了模型名称）
	if parsedModel.ModelName != "" && parsedModel.ModelName != existingModel.ModelName {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			fmt.Sprintf(
				"Schema model name mismatch: schema has '%s', model has '%s'",
				parsedModel.ModelName,
				existingModel.ModelName,
			),
		)
	}

	// 4. 比较字段差异
	existingFieldNames := make(map[string]*modeldesign.FieldDefinition)
	for _, field := range existingModel.Fields {
		existingFieldNames[field.Name] = field
	}

	schemaFieldNames := make(map[string]*modeldesign.FieldDefinition)
	for _, field := range parsedModel.Fields {
		schemaFieldNames[field.Name] = field
	}

	// 5. 识别需要添加的字段
	var fieldsToAdd []*modeldesign.FieldDefinition
	var skippedFields []string
	for _, schemaField := range parsedModel.Fields {
		if existingField, exists := existingFieldNames[schemaField.Name]; exists {
			// 字段已存在，检查类型是否冲突
			if existingField.Type.Format != schemaField.Type.Format {
				return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid,
					fmt.Sprintf("Field '%s' type conflict: existing type %s, schema type %s",
						schemaField.Name, existingField.Type.Format, schemaField.Type.Format))
			}
			skippedFields = append(skippedFields, schemaField.Name)
		} else {
			// 新字段，需要添加
			schemaField.ModelID = modelID
			fieldsToAdd = append(fieldsToAdd, schemaField)
		}
	}

	result := &SyncModelSchemaResult{
		Model:         existingModel,
		FieldsAdded:   0,
		FieldsSkipped: skippedFields,
		FieldsDeleted: 0,
		DeletedFields: []string{},
	}

	// 6. 添加新字段
	if len(fieldsToAdd) > 0 {
		addCmd := AddFieldCommand{
			ModelID: modelID,
			Fields:  fieldsToAdd,
		}

		if err := s.AddFieldSync(ctx, addCmd); err != nil {
			return nil, bizerrors.Wrapf(err, "Failed to add fields")
		}
		result.FieldsAdded = len(fieldsToAdd)
		logger.Infof(ctx, "Added %d fields to model %s", len(fieldsToAdd), modelID)
	}

	// 7. 删除额外字段（如果启用）
	if deleteExtraFields {
		deletedFields, err := s.deleteExtraFields(ctx, modelID, existingModel, schemaFieldNames, logger)
		if err != nil {
			return nil, err
		}
		result.FieldsDeleted = len(deletedFields)
		result.DeletedFields = deletedFields
	}

	// 8. 重新获取更新后的模型
	updatedModel, err := s.modelRepo.GetByID(ctx, modelID, opts)
	if err != nil {
		return nil, err
	}
	result.Model = updatedModel

	logger.Infof(
		ctx, "Successfully synced model schema: %s (added: %d, deleted: %d)",
		modelID,
		result.FieldsAdded,
		result.FieldsDeleted,
	)
	return result, nil
}

func (s *ModelDesignAppService) collectFieldsToDelete(
	ctx context.Context,
	existingModel *modeldesign.DataModel,
	schemaFieldNames map[string]*modeldesign.FieldDefinition,
) ([]string, error) {
	systemFieldNames := map[string]bool{
		"id":        true,
		"createdAt": true,
		"updatedAt": true,
	}

	fieldsToDelete := make([]string, 0, len(existingModel.Fields))
	for _, existingField := range existingModel.Fields {
		// 跳过系统字段
		if systemFieldNames[existingField.Name] {
			continue
		}

		// 如果字段不在schema中，且不是系统字段
		if _, inSchema := schemaFieldNames[existingField.Name]; inSchema {
			continue
		}

		fieldsToDelete = append(fieldsToDelete, existingField.Name)
	}

	return fieldsToDelete, nil
}

func (s *ModelDesignAppService) deleteExtraFields(
	ctx context.Context,
	modelID string,
	existingModel *modeldesign.DataModel,
	schemaFieldNames map[string]*modeldesign.FieldDefinition,
	logger logfacade.Logger,
) ([]string, error) {
	fieldsToDelete, err := s.collectFieldsToDelete(ctx, existingModel, schemaFieldNames)
	if err != nil {
		return nil, err
	}
	if len(fieldsToDelete) == 0 {
		return nil, nil
	}

	for _, fieldName := range fieldsToDelete {
		removeCmd := RemoveFieldCommand{
			ModelID:   modelID,
			FieldName: fieldName,
		}
		if err := s.RemoveFieldSync(ctx, removeCmd); err != nil {
			return nil, bizerrors.Wrapf(err, "Failed to delete field '%s'", fieldName)
		}
	}

	logger.Infof(ctx, "Deleted %d fields from model %s", len(fieldsToDelete), modelID)
	return fieldsToDelete, nil
}

// assignDisplayOrders computes and assigns a display_order for each field by appending
// them sequentially after the current tail order in the model. This ensures the new
// fields appear last in the ordered list and maintains strict lexicographic ordering.
func (s *ModelDesignAppService) assignDisplayOrders(
	ctx context.Context,
	modelID string,
	fields []*modeldesign.FieldDefinition,
) error {
	tail, err := s.modelRepo.GetTailFieldDisplayOrder(ctx, modelID)
	if err != nil {
		return bizerrors.Wrapf(err, "get tail field display order for model %s", modelID)
	}
	prev := tail
	for _, field := range fields {
		order, err := lexorder.Midpoint(prev, "")
		if err != nil {
			return bizerrors.Wrapf(err, "compute display order for field %s", field.Name)
		}
		field.DisplayOrder = order
		prev = order
	}
	return nil
}
