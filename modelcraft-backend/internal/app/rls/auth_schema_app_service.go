package rls

import (
	"context"

	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/rls"
	"modelcraft/pkg/bizerrors"
)

// AuthSchemaRepository 定义 AuthSchema 仓储接口
type AuthSchemaRepository interface {
	GetByProjectID(ctx context.Context, orgName, projectSlug string) (*rls.AuthSchema, error)
	Save(ctx context.Context, orgName, projectSlug string, authSchema *rls.AuthSchema) error
	DeleteByProjectID(ctx context.Context, orgName, projectSlug string) error
}

// AuthSchemaAppService AuthSchema 应用服务
type AuthSchemaAppService struct {
	authSchemaRepo AuthSchemaRepository
	projectRepo    project.ProjectRepository
}

// NewAuthSchemaAppService 创建 AuthSchemaAppService
func NewAuthSchemaAppService(
	authSchemaRepo AuthSchemaRepository,
	projectRepo project.ProjectRepository,
) *AuthSchemaAppService {
	return &AuthSchemaAppService{
		authSchemaRepo: authSchemaRepo,
		projectRepo:    projectRepo,
	}
}

// SetAuthSchema 设置 Project AuthSchema
func (s *AuthSchemaAppService) SetAuthSchema(ctx context.Context, orgName string,
	input SetProjectAuthSchemaInput) (*rls.AuthSchema, error) {

	// 1. 检查 Project 是否存在
	p, err := s.projectRepo.GetByNameAndOrg(ctx, input.ProjectSlug, orgName)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ProjectNotFound, input.ProjectSlug)
	}

	// 2. 校验变量（不允许声明 uid）
	for _, v := range input.Variables {
		if v.Name == "uid" {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid,
				"'uid' is a reserved variable and cannot be declared")
		}
	}

	// 3. 保存 AuthSchema
	authSchema := &rls.AuthSchema{
		ProjectID: input.ProjectSlug,
		Variables: input.Variables,
	}

	if err := s.authSchemaRepo.Save(ctx, orgName, input.ProjectSlug, authSchema); err != nil {
		return nil, err
	}

	return authSchema, nil
}

// GetAuthSchema 获取 Project AuthSchema
func (s *AuthSchemaAppService) GetAuthSchema(ctx context.Context, orgName, projectSlug string) (*rls.AuthSchema, error) {
	authSchema, err := s.authSchemaRepo.GetByProjectID(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	if authSchema == nil {
		// 返回空 AuthSchema
		return &rls.AuthSchema{ProjectID: projectSlug, Variables: []rls.AuthVariable{}}, nil
	}
	return authSchema, nil
}
