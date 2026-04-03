package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/infrastructure/dbgen"

	domainauth "modelcraft/internal/domain/auth"
)

// SqlSecurityAuditLogRepository is the sqlc-based implementation of auth.SecurityAuditLogRepository.
type SqlSecurityAuditLogRepository struct {
	q dbgen.Querier
}

// NewSqlSecurityAuditLogRepository creates a SqlSecurityAuditLogRepository.
func NewSqlSecurityAuditLogRepository(q dbgen.Querier) domainauth.SecurityAuditLogRepository {
	return &SqlSecurityAuditLogRepository{q: q}
}

// Insert inserts a security audit log record.
func (r *SqlSecurityAuditLogRepository) Insert(ctx context.Context, log *domainauth.SecurityAuditLog) error {
	var detail *json.RawMessage
	if log.Detail != nil {
		bytes, err := json.Marshal(log.Detail)
		if err != nil {
			return fmt.Errorf("marshal security audit detail: %w", err)
		}
		raw := json.RawMessage(bytes)
		detail = &raw
	}

	return ExecWithErrorHandling(func() error {
		return r.q.InsertSecurityAuditLog(ctx, dbgen.InsertSecurityAuditLogParams{
			ID:     log.ID,
			UserID: log.UserID,
			Event:  string(log.Event),
			Detail: detail,
		})
	})
}

var _ domainauth.SecurityAuditLogRepository = (*SqlSecurityAuditLogRepository)(nil)
