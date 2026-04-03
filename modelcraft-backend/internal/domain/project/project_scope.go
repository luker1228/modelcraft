package project

import (
	bizerrors "modelcraft/pkg/bizerrors"
	"strings"
)

// ProjectScope 项目作用域，表示某个组织下的某个项目。
// 用于需要完整项目上下文的实体，确保 OrgName 和 ProjectSlug 均不为空。
type ProjectScope struct {
	OrgName     string
	ProjectSlug string
}

// Validate 验证项目作用域的必填字段
func (s *ProjectScope) Validate() error {
	if strings.TrimSpace(s.OrgName) == "" {
		return bizerrors.Errorf("OrgName cant be blank")
	}
	if strings.TrimSpace(s.ProjectSlug) == "" {
		return bizerrors.Errorf("ProjectSlug cant be blank")
	}
	return nil
}

// NewProjectScope 创建项目作用域并验证必填字段
func NewProjectScope(orgName, projectSlug string) (ProjectScope, error) {
	s := ProjectScope{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	}
	if err := s.Validate(); err != nil {
		return s, err
	}
	return s, nil
}

// GetFullPath 获取项目完整路径
// 返回格式: orgName.projectSlug
func (s *ProjectScope) GetFullPath() string {
	return s.OrgName + "." + s.ProjectSlug
}
