package modeldatabase

import "context"

// ModelDatabaseRepository 项目数据库注册仓储接口
type ModelDatabaseRepository interface {
	// Create 创建数据库注册记录
	Create(ctx context.Context, db *ModelDatabase) error
	// GetByID 按 ID 查询（需 orgName + projectSlug 隔离）
	GetByID(ctx context.Context, orgName, projectSlug, id string) (*ModelDatabase, error)
	// GetByName 按数据库名查询
	GetByName(ctx context.Context, orgName, projectSlug, name string) (*ModelDatabase, error)
	// List 列出项目下所有已注册的数据库
	List(ctx context.Context, orgName, projectSlug string) ([]*ModelDatabase, error)
	// Update 更新数据库注册记录
	Update(ctx context.Context, orgName, projectSlug string, db *ModelDatabase) error
	// Delete 软删除数据库注册记录
	Delete(ctx context.Context, orgName, projectSlug, id string) error
}
