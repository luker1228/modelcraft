package project

import "modelcraft/internal/domain/project"

// ============================================================================
// Project Commands
// ============================================================================

// CreateProjectCommand 创建项目命令（同时原子创建关联的 cluster）
type CreateProjectCommand struct {
	OrgName            string                       // 组织名称
	Slug               string                       // 项目标识符（组织内唯一）
	Title              string                       // 显示名称
	Description        string                       // 描述
	LoginURL           string                       // 登录URL（可选）
	ClusterInput       CreateClusterForProjectInput // 必填：随项目一起创建的 cluster 信息
	SkipConnectionTest bool                         // 是否跳过连接验证（默认 false）
}

// CreateClusterForProjectInput 创建项目时内嵌的 cluster 连接信息
type CreateClusterForProjectInput struct {
	Title             string // cluster 显示名称
	Description       string // cluster 描述（可选）
	Host              string
	Port              int
	Username          string
	Password          string
	ConnectionTimeout int
}

// UpdateProjectCommand 更新项目命令
type UpdateProjectCommand struct {
	OrgName     string // 组织名称
	Slug        string // 项目标识符（用于查找）
	Title       string // 新的显示名称（空字符串表示不更新）
	Description string // 新的描述（空字符串表示不更新）
	LoginURL    string // 新的登录URL（空字符串表示不更新）
}

// DeleteProjectCommand 删除项目命令（实际是归档，同时级联删除 cluster）
type DeleteProjectCommand struct {
	OrgName string // 组织名称
	Slug    string // 项目标识符
}

// GetProjectCommand 获取项目命令
type GetProjectCommand struct {
	OrgName string // 组织名称
	Slug    string // 项目标识符
}

// ListProjectsCommand 列举项目命令
type ListProjectsCommand struct {
	OrgName string                 // 组织名称
	Status  *project.ProjectStatus // 可选的状态过滤
}
