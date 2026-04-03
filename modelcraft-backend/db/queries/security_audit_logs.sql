-- name: InsertSecurityAuditLog :exec
INSERT INTO security_audit_logs (id, user_id, event, detail, created_at)
VALUES (?, ?, ?, ?, NOW());
