package rbac

import "time"

// SnapshotPermissionEntry 快照中的权限点条目（存储 ID 引用，不复制权限点完整数据）
type SnapshotPermissionEntry struct {
	PermissionID string
	SortOrder    int
}

// BundleSnapshot 权限包历史快照（值对象，只读）
// 每次权限列表变更时自动创建，最多保留最近 5 个历史版本
type BundleSnapshot struct {
	ID           string
	BundleID     string
	Version      int
	Permissions  []SnapshotPermissionEntry
	CreatedAt    time.Time
	CreatedBy    *string
	RestoredFrom *int // 若为回滚操作，指向来源版本号；否则为 nil
}
