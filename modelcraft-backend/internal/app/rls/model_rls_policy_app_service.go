package rls

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/rls"
	"modelcraft/pkg/bizerrors"
)

// ModelRLSPolicyAppService RLS 策略应用服务
type ModelRLSPolicyAppService struct {
	policyRepo     modeldesign.ModelRLSPolicyRepository
	modelRepo      modeldesign.ModelRepository
	authSchemaRepo AuthSchemaRepository
	validator      rls.PolicyValidator
}

// NewModelRLSPolicyAppService 创建 ModelRLSPolicyAppService
func NewModelRLSPolicyAppService(
	policyRepo modeldesign.ModelRLSPolicyRepository,
	modelRepo modeldesign.ModelRepository,
	authSchemaRepo AuthSchemaRepository,
	validator rls.PolicyValidator,
) *ModelRLSPolicyAppService {
	return &ModelRLSPolicyAppService{
		policyRepo:     policyRepo,
		modelRepo:      modelRepo,
		authSchemaRepo: authSchemaRepo,
		validator:      validator,
	}
}

// SetPolicy 设置 Model RLS 策略
func (s *ModelRLSPolicyAppService) SetPolicy(ctx context.Context, orgName, projectSlug string,
	input SetModelRLSPolicyInput,
) (*modeldesign.ModelRLSPolicy, error) {
	// 1. 检查 Model 是否存在并获取字段
	model, err := s.modelRepo.GetByID(ctx, input.ModelID)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, input.ModelID)
	}

	// 2. 检查是否有 owner 字段（EndUserRef 类型）
	hasOwner := false
	for _, f := range model.Fields {
		if f.IsEndUserRef() {
			hasOwner = true
			break
		}
	}
	if !hasOwner {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid,
			"Model has no owner field (EndUserRef), cannot set RLS policy")
	}

	// 3. 获取 AuthSchema 用于校验
	authSchema, err := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	if authSchema == nil {
		authSchema = &rls.AuthSchema{ProjectID: projectSlug}
	}

	// 4. 校验五件套表达式
	exprs := map[rls.ExprType]rls.JsonExpr{
		rls.ExprTypeSelectPredicate: input.SelectPredicate,
		rls.ExprTypeInsertCheck:     input.InsertCheck,
		rls.ExprTypeUpdatePredicate: input.UpdatePredicate,
		rls.ExprTypeUpdateCheck:     input.UpdateCheck,
		rls.ExprTypeDeletePredicate: input.DeletePredicate,
	}

	for exprType, expr := range exprs {
		if errors := s.validator.Validate(ctx, expr, exprType, model, authSchema); len(errors) > 0 {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid,
				errors[0].Path+": "+errors[0].Message)
		}
	}

	// 5. 保存策略
	policy := &modeldesign.ModelRLSPolicy{
		ModelID:         input.ModelID,
		SelectPredicate: input.SelectPredicate,
		InsertCheck:     input.InsertCheck,
		UpdatePredicate: input.UpdatePredicate,
		UpdateCheck:     input.UpdateCheck,
		DeletePredicate: input.DeletePredicate,
	}

	if err := s.policyRepo.Save(ctx, orgName, projectSlug, policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// GetPolicy 获取 Model RLS 策略
func (s *ModelRLSPolicyAppService) GetPolicy(
	ctx context.Context, orgName, projectSlug, modelID string,
) (*modeldesign.ModelRLSPolicy, error) {
	return s.policyRepo.GetByModelID(ctx, orgName, projectSlug, modelID)
}

// ApplyPreset 应用预设策略
func (s *ModelRLSPolicyAppService) ApplyPreset(ctx context.Context, orgName, projectSlug, modelID string,
	preset rls.RLSPreset,
) (*modeldesign.ModelRLSPolicy, error) {
	policy := &modeldesign.ModelRLSPolicy{ModelID: modelID}
	policy.ApplyPreset(preset)

	input := SetModelRLSPolicyInput{
		ModelID:         modelID,
		SelectPredicate: policy.SelectPredicate,
		InsertCheck:     policy.InsertCheck,
		UpdatePredicate: policy.UpdatePredicate,
		UpdateCheck:     policy.UpdateCheck,
		DeletePredicate: policy.DeletePredicate,
	}

	return s.SetPolicy(ctx, orgName, projectSlug, input)
}

// ValidateExpr 校验 RLS 表达式（用于 UI 实时校验）
func (s *ModelRLSPolicyAppService) ValidateExpr(ctx context.Context, orgName, projectSlug, modelID string,
	exprType rls.ExprType, expr rls.JsonExpr,
) []rls.ValidationError {
	// 获取 Model 和 AuthSchema
	model, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil || model == nil {
		return []rls.ValidationError{{
			Path:    "modelId",
			Message: "Model not found",
			Code:    "MODEL_NOT_FOUND",
		}}
	}

	authSchema, _ := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
	if authSchema == nil {
		authSchema = &rls.AuthSchema{ProjectID: projectSlug}
	}

	return s.validator.Validate(ctx, expr, exprType, model, authSchema)
}
