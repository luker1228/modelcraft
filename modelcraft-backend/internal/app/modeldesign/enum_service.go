package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizutils"
	"time"

	domainProject "modelcraft/internal/domain/project"

	bizerrors "modelcraft/pkg/bizerrors"
)

// EnumAppService 枚举应用服务
type EnumAppService struct {
	enumRepo    modeldesign.EnumRepository
	projectRepo domainProject.ProjectRepository
}

// NewEnumAppService 创建枚举服务实例
func NewEnumAppService(
	enumRepo modeldesign.EnumRepository,
	projectRepo domainProject.ProjectRepository,
) *EnumAppService {
	return &EnumAppService{
		enumRepo:    enumRepo,
		projectRepo: projectRepo,
	}
}

func (s *EnumAppService) ensureProjectExists(projectSlug, orgName string) error {
	if s.projectRepo == nil {
		return nil
	}

	exists, err := s.projectRepo.ExistsByName(context.Background(), projectSlug, orgName)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to check project existence")
	}
	if !exists {
		return bizerrors.NewError(bizerrors.ProjectNotFound, projectSlug)
	}
	return nil
}

// CreateEnum 创建枚举定义
func (s *EnumAppService) CreateEnum(ctx context.Context, cmd CreateEnumCommand) (*modeldesign.EnumDefinition, error) {
	orgName := cmd.OrgName
	if orgName == "" {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return nil, err
	}

	// 生成ID
	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to generate enum ID")
	}

	// 创建枚举定义
	enum := &modeldesign.EnumDefinition{
		ID: id,
		ProjectScope: domainProject.ProjectScope{
			OrgName:     orgName,
			ProjectSlug: cmd.ProjectSlug,
		},
		Name:          cmd.Name,
		DisplayName:   cmd.DisplayName,
		Description:   cmd.Description,
		Options:       cmd.Options,
		IsMultiSelect: cmd.IsMultiSelect,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 验证
	if err := enum.Validate(); err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, err.Error())
	}

	// 保存
	if err := s.enumRepo.Create(enum); err != nil {
		// Check if it's a duplicate key error (enum already exists)
		if shared.IsRepoError(err, shared.ErrTypeDuplicatedKey) {
			return nil, bizerrors.NewError(bizerrors.EnumAlreadyExists, cmd.Name)
		}
		return nil, bizerrors.Wrapf(err, "failed to create enum")
	}

	return enum, nil
}

// UpdateEnum 更新枚举定义
func (s *EnumAppService) UpdateEnum(ctx context.Context, cmd UpdateEnumCommand) error {
	orgName := cmd.OrgName
	if orgName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return err
	}

	// 查找现有枚举
	enum, err := s.enumRepo.FindByName(orgName, cmd.ProjectSlug, cmd.Name)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to find enum")
	}
	if enum == nil {
		return bizerrors.NewError(bizerrors.EnumNotFound, cmd.Name)
	}

	// 更新
	if err := enum.Update(cmd.DisplayName, cmd.Description, cmd.Options); err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, err.Error())
	}

	// 保存
	if err := s.enumRepo.Update(enum); err != nil {
		return bizerrors.Wrapf(err, "failed to save enum")
	}

	return nil
}

// DeleteEnum 删除枚举定义
func (s *EnumAppService) DeleteEnum(ctx context.Context, cmd DeleteEnumCommand) error {
	orgName := cmd.OrgName
	if orgName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return err
	}

	// 检查是否被引用
	isReferenced, fieldNames, err := s.enumRepo.IsReferencedByFields(orgName, cmd.ProjectSlug, cmd.Name)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to check enum references")
	}
	if isReferenced {
		fieldNamesStr := bizutils.MarshalToStringIgnoreErr(fieldNames)
		return bizerrors.NewError(bizerrors.CannotDeleteReferencedEnum, cmd.Name, fieldNamesStr)
	}

	// 删除
	if err := s.enumRepo.Delete(orgName, cmd.ProjectSlug, cmd.Name); err != nil {
		return bizerrors.Wrapf(err, "failed to delete enum")
	}

	return nil
}

// GetEnum 获取枚举定义
func (s *EnumAppService) GetEnum(ctx context.Context, cmd GetEnumCommand) (*modeldesign.EnumDefinition, error) {
	orgName := cmd.OrgName
	if orgName == "" {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return nil, err
	}

	enum, err := s.enumRepo.FindByName(orgName, cmd.ProjectSlug, cmd.Name)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to find enum")
	}
	if enum == nil {
		return nil, bizerrors.NewError(bizerrors.EnumNotFound, cmd.Name)
	}
	return enum, nil
}

// ListEnums 列出所有枚举定义
func (s *EnumAppService) ListEnums(ctx context.Context, cmd ListEnumsCommand) ([]*modeldesign.EnumDefinition, error) {
	orgName := cmd.OrgName
	if orgName == "" {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return nil, err
	}

	enums, err := s.enumRepo.List(orgName, cmd.ProjectSlug)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list enums")
	}
	return enums, nil
}

// GetEnumReferences 获取枚举的引用字段列表
func (s *EnumAppService) GetEnumReferences(ctx context.Context, cmd GetEnumReferencesCommand) ([]string, error) {
	orgName := cmd.OrgName
	if orgName == "" {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return nil, err
	}

	isReferenced, fieldNames, err := s.enumRepo.IsReferencedByFields(orgName, cmd.ProjectSlug, cmd.Name)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get enum references")
	}
	if !isReferenced {
		return []string{}, nil
	}
	return fieldNames, nil
}

// ValidateEnumCodes 验证枚举code值
func (s *EnumAppService) ValidateEnumCodes(ctx context.Context, cmd ValidateEnumCodesCommand) error {
	orgName := cmd.OrgName
	if orgName == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName is required")
	}

	if err := s.ensureProjectExists(cmd.ProjectSlug, orgName); err != nil {
		return err
	}

	enum, err := s.enumRepo.FindByName(orgName, cmd.ProjectSlug, cmd.EnumName)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to find enum")
	}
	if enum == nil {
		return bizerrors.NewError(bizerrors.EnumNotFound, cmd.EnumName)
	}

	return enum.ValidateCodes(cmd.Codes)
}
