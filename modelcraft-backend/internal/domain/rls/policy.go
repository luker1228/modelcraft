package rls

import (
	"errors"
	"time"
)

// ErrNoMatchingPolicy is returned when no RLS policy matches the request.
var ErrNoMatchingPolicy = errors.New("RLS deny: no matching policy")

// ErrPolicyNotFound is returned when a specific policy lookup yields no result.
var ErrPolicyNotFound = errors.New("RLS policy not found")

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
