package modeldesign

import (
	"time"
)

// DeploymentStatus 部署状态枚举
type DeploymentStatus string

const (
	DeploymentSynced  DeploymentStatus = "synced"  // 已同步
	DeploymentPending DeploymentStatus = "pending" // 待同步
	DeploymentFailed  DeploymentStatus = "failed"  // 同步失败
)

// ModelChangeType 模型变更类型
type ModelChangeType string

const (
	ChangeCreateModel ModelChangeType = "create_model"
	ChangeDropModel   ModelChangeType = "drop_model" // 删除模型
	ChangeRemoveModel ModelChangeType = "remove_model"
	ChangeAddField    ModelChangeType = "add_field"
	ChangeUpdateField ModelChangeType = "update_field"
	ChangeRemoveField ModelChangeType = "remove_field"
	ChangeUpdateModel ModelChangeType = "update_model"
)

// ModelChange 模型变更请求
type ModelChange struct {
	Type        ModelChangeType `json:"type"`
	ModelID     string          `json:"modelId"`
	Changes     interface{}     `json:"changes"`
	RequiresDDL bool            `json:"requiresDDL"`
}

// ChangeResult 模型变更结果
type ChangeResult struct {
	ChangeID                string          `json:"changeId"`
	Type                    ModelChangeType `json:"type"`
	Success                 bool            `json:"success"`
	PlatformSuccess         bool            `json:"platformSuccess"`
	ClientDeploymentNeeded  bool            `json:"clientDeploymentNeeded"`
	ClientDeploymentSuccess bool            `json:"clientDeploymentSuccess"`
	DeploymentError         string          `json:"deploymentError,omitempty"`
	CanRetryDeployment      bool            `json:"canRetryDeployment"`
}

// DeploymentHistory 部署历史记录
type DeploymentHistory struct {
	ID            string          `json:"id" db:"id"`
	ModelID       string          `json:"modelId" db:"model_id"`
	ChangeType    ModelChangeType `json:"changeType" db:"change_type"`
	DDLStatements []string        `json:"ddlStatements" db:"ddl_statements"`
	Status        string          `json:"status" db:"status"`
	Error         string          `json:"error,omitempty" db:"error"`
	CreatedAt     time.Time       `json:"createdAt" db:"created_at"`
	CompletedAt   *time.Time      `json:"completedAt,omitempty" db:"completed_at"`
}
