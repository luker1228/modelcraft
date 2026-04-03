package auth

import (
	"context"
	"time"
)

// SecurityAuditEvent represents event type for security audit logs.
type SecurityAuditEvent string

const (
	EventReuseDetected SecurityAuditEvent = "REUSE_DETECTED"
)

// SecurityAuditLog represents a security-related audit record.
type SecurityAuditLog struct {
	ID        string
	UserID    string
	Event     SecurityAuditEvent
	Detail    map[string]any
	CreatedAt time.Time
}

// SecurityAuditLogRepository defines security audit log persistence operations.
type SecurityAuditLogRepository interface {
	Insert(ctx context.Context, log *SecurityAuditLog) error
}
