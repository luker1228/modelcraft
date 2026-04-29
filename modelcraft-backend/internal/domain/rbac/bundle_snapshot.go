package rbac

import "time"

// SnapshotPermissionEntry 快照中的旧权限点条目（兼容字段）。
type SnapshotPermissionEntry struct {
	PermissionID string
	SortOrder    int
}

// SnapshotItemEntry 快照中的数据权限 item 条目（新结构）。
type SnapshotItemEntry struct {
	ModelID            string
	GrantType          PermissionType
	Preset             *PermissionPreset
	CustomPermissionID *string
	SortOrder          int
}

// BundleSnapshot 权限包历史快照（值对象，只读）
// 每次权限列表变更时自动创建，最多保留最近 5 个历史版本
type BundleSnapshot struct {
	ID       string
	BundleID string
	Version  int
	// Permissions 旧字段（兼容）。
	Permissions []SnapshotPermissionEntry
	// Items 新字段。
	Items        []SnapshotItemEntry
	CreatedAt    time.Time
	CreatedBy    *string
	RestoredFrom *int // 若为回滚操作，指向来源版本号；否则为 nil
}
