package rls

// RLSPreset RLS 预设策略
type RLSPreset string

const (
	RLSPresetReadWriteOwner    RLSPreset = "READ_WRITE_OWNER"     // 默认：读写自己
	RLSPresetReadAllWriteOwner RLSPreset = "READ_ALL_WRITE_OWNER" // 读取全部，写自己
	RLSPresetReadAll           RLSPreset = "READ_ALL"             // 只读全部
	RLSPresetReadWriteAll      RLSPreset = "READ_WRITE_ALL"       // 读写全部（高危）
	RLSPresetNoAccess          RLSPreset = "NO_ACCESS"            // 无访问
)

// IsDangerous 判断是否为高危策略
func (p RLSPreset) IsDangerous() bool {
	return p == RLSPresetReadWriteAll
}

// String 返回显示名称
func (p RLSPreset) String() string {
	return string(p)
}
