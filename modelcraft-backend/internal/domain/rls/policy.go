package rls

import "time"

// Action 操作类型
type Action string

const (
	ActionRead   Action = "read"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Policy 单条 RLS 策略
type Policy struct {
	ID            int64     `json:"id"`
	OrgName       string    `json:"orgName"`
	ProjectSlug   string    `json:"projectSlug"`
	ModelID       string    `json:"modelId"`
	PolicyName    string    `json:"policyName"`
	Action        Action    `json:"action"`
	Role          string    `json:"role"`
	UsingExpr     JsonExpr  `json:"usingExpr"`
	WithCheckExpr JsonExpr  `json:"withCheckExpr"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
