package modeldatabase

import "time"

// DatabaseMode 数据库管理模式
type DatabaseMode string

const (
	// DatabaseModeSelfHosted 自托管模式（用户自行管理）
	DatabaseModeSelfHosted DatabaseMode = "self_hosted"
	// DatabaseModeManaged 托管模式（平台管理）
	DatabaseModeManaged DatabaseMode = "managed"
)

// ModelDatabase 项目数据库注册实体
type ModelDatabase struct {
	ID              string
	OrgName         string
	ProjectSlug     string
	ClusterID       string
	Name            string
	Title           string
	Description     string
	Mode            DatabaseMode
	LatestSyncJobID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
