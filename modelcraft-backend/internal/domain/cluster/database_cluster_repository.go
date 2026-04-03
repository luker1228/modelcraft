package cluster

import (
	"context"
	"time"
)

// DatabaseClusterRepository 数据库集群仓储接口
type DatabaseClusterRepository interface {
	// Create 创建数据库集群
	Create(ctx context.Context, cluster *DatabaseCluster) error

	// Update 更新数据库集群（org + project scoped，防止跨租户访问）
	Update(ctx context.Context, orgName, projectSlug string, cluster *DatabaseCluster) error

	// GetByID 根据主键ID获取数据库集群（org scoped，防止跨租户访问）
	GetByID(ctx context.Context, orgName, id string) (*DatabaseCluster, error)

	// GetByProjectKey 根据组织和项目名称获取集群（用于一对一关系查询）
	GetByProjectKey(ctx context.Context, orgName, projectSlug string) (*DatabaseCluster, error)

	// List 列出项目下的数据库集群（org + project scoped）
	List(ctx context.Context, orgName, projectSlug string, status ...ClusterStatus) ([]*DatabaseCluster, error)

	// Delete 删除数据库集群（软删除）（org + project scoped）
	Delete(ctx context.Context, orgName, projectSlug, id string) error

	// ExistsByProjectKey 检查项目是否已有集群（用于一对一约束验证）
	ExistsByProjectKey(ctx context.Context, orgName, projectSlug string) (bool, error)

	// ListUpdatedAfter 获取项目下指定时间之后更新的集群列表（org + project scoped）
	ListUpdatedAfter(
		ctx context.Context,
		orgName string,
		projectSlug string,
		updatedAfter time.Time,
		status ...ClusterStatus,
	) ([]*DatabaseCluster, error)
}
