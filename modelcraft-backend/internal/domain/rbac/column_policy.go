package rbac

// ColumnAccessMode 列访问模式枚举（对齐 API 合约）
type ColumnAccessMode string

const (
	ColumnAccessModeVisible  ColumnAccessMode = "VISIBLE"  // 可见可编辑（完整访问）
	ColumnAccessModeReadonly ColumnAccessMode = "READONLY" // 可见但不可编辑
	ColumnAccessModeMasked   ColumnAccessMode = "MASKED"   // 脱敏显示
	ColumnAccessModeHidden   ColumnAccessMode = "HIDDEN"   // 完全隐藏
)

// columnAccessModeOrder 列访问模式宽泛度排序（值越大权限越宽）
var columnAccessModeOrder = map[ColumnAccessMode]int{
	ColumnAccessModeHidden:   1,
	ColumnAccessModeMasked:   2,
	ColumnAccessModeReadonly: 3,
	ColumnAccessModeVisible:  4,
}

// IsValid 判断枚举值是否合法
func (m ColumnAccessMode) IsValid() bool {
	_, ok := columnAccessModeOrder[m]
	return ok
}

// ColumnRule 单个字段的列访问规则
type ColumnRule struct {
	FieldName   string           `json:"field_name"`
	Mode        ColumnAccessMode `json:"mode"`
	MaskPattern string           `json:"mask_pattern,omitempty"`
}

// ColumnPolicy 列策略（对齐 API 合约结构）
// nil 表示全列默认（按 DefaultMode 决定，默认 VISIBLE）
type ColumnPolicy struct {
	DefaultMode ColumnAccessMode `json:"default_mode"`
	Rules       []ColumnRule     `json:"rules"`
}

// mergeColumnPolicy 合并两个列策略，取更宽泛（更高权限）的结果
// 规则：VISIBLE > READONLY > MASKED > HIDDEN
// - DefaultMode 取两者中更宽泛的
// - Rules 中同 fieldName 的条目取更宽泛的 Mode；仅一方有的条目直接保留
func mergeColumnPolicy(a, b *ColumnPolicy) *ColumnPolicy {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	// DefaultMode 取更宽泛
	mergedDefault := a.DefaultMode
	if columnAccessModeOrder[b.DefaultMode] > columnAccessModeOrder[a.DefaultMode] {
		mergedDefault = b.DefaultMode
	}

	// 合并 Rules：按 fieldName 索引，取更宽泛的 Mode
	ruleMap := make(map[string]ColumnRule)
	for _, r := range a.Rules {
		ruleMap[r.FieldName] = r
	}
	for _, r := range b.Rules {
		if existing, ok := ruleMap[r.FieldName]; ok {
			if columnAccessModeOrder[r.Mode] > columnAccessModeOrder[existing.Mode] {
				ruleMap[r.FieldName] = r
			}
		} else {
			ruleMap[r.FieldName] = r
		}
	}

	mergedRules := make([]ColumnRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		mergedRules = append(mergedRules, r)
	}

	return &ColumnPolicy{
		DefaultMode: mergedDefault,
		Rules:       mergedRules,
	}
}
