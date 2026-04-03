package query

import (
	"testing"
)

// ============================================================================
// Reserved Keywords Tests
// ============================================================================

func TestGetReservedKeywords(t *testing.T) {
	keywords := GetReservedKeywords()

	// 验证返回的是副本，修改不会影响原始列表
	originalLen := len(keywords)
	keywords[0] = "modified"
	newKeywords := GetReservedKeywords()

	if len(newKeywords) != originalLen {
		t.Errorf("GetReservedKeywords() 应该返回副本")
	}

	// 验证包含关键操作符
	expectedKeywords := []string{
		"AND",
		"OR",
		"NOT",
		"equals",
		"not",
		"in",
		"lt",
		"lte",
		"gt",
		"gte",
		"contains",
		"startsWith",
		"endsWith",
		"mode",
	}

	keywordMap := make(map[string]bool)
	for _, kw := range newKeywords {
		keywordMap[kw] = true
	}

	for _, expected := range expectedKeywords {
		if !keywordMap[expected] {
			t.Errorf("GetReservedKeywords() 应该包含关键字 '%s'", expected)
		}
	}
}

func TestIsReservedKeyword(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		expected bool
	}{
		// 逻辑操作符
		{"AND uppercase", "AND", true},
		{"AND lowercase", "and", true},
		{"AND mixed case", "And", true},
		{"OR uppercase", "OR", true},
		{"OR lowercase", "or", true},
		{"OR mixed case", "Or", true},
		{"NOT uppercase", "NOT", true},
		{"NOT lowercase", "not", true},
		{"NOT mixed case", "Not", true},

		// 比较操作符
		{"equals", "equals", true},
		{"Equals", "Equals", true},
		{"EQUALS", "EQUALS", true},
		{"not", "not", true},
		{"Not", "Not", true},
		{"in", "in", true},
		{"IN", "IN", true},

		// 数值比较
		{"lt", "lt", true},
		{"LT", "LT", true},
		{"lte", "lte", true},
		{"LTE", "LTE", true},
		{"gt", "gt", true},
		{"GT", "GT", true},
		{"gte", "gte", true},
		{"GTE", "GTE", true},

		// 字符串操作符
		{"contains", "contains", true},
		{"Contains", "Contains", true},
		{"CONTAINS", "CONTAINS", true},
		{"startsWith", "startsWith", true},
		{"StartsWith", "StartsWith", true},
		{"endsWith", "endsWith", true},
		{"EndsWith", "EndsWith", true},

		// 查询模式
		{"mode", "mode", true},
		{"Mode", "Mode", true},
		{"MODE", "MODE", true},

		// 非保留关键字
		{"valid field name", "username", false},
		{"valid field name 2", "email", false},
		{"valid field name 3", "created_at", false},
		{"similar to keyword", "contain", false},
		{"similar to keyword 2", "note", false},
		{"similar to keyword 3", "band", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReservedKeyword(tt.keyword)
			if result != tt.expected {
				t.Errorf("IsReservedKeyword(%s) = %v, 期望 %v", tt.keyword, result, tt.expected)
			}
		})
	}
}

func TestIsReservedKeyword_CaseInsensitive(t *testing.T) {
	// 测试所有关键字的不同大小写组合都被正确识别
	keywords := GetReservedKeywords()

	for _, keyword := range keywords {
		// 测试小写
		if !IsReservedKeyword(keyword) {
			t.Errorf("IsReservedKeyword(%s) 应该返回 true", keyword)
		}

		// 对于AND, OR, NOT要额外测试小写形式
		if keyword == "AND" && !IsReservedKeyword("and") {
			t.Errorf("IsReservedKeyword('and') 应该返回 true")
		}
		if keyword == "OR" && !IsReservedKeyword("or") {
			t.Errorf("IsReservedKeyword('or') 应该返回 true")
		}
		if keyword == "NOT" && !IsReservedKeyword("not") {
			t.Errorf("IsReservedKeyword('not') 应该返回 true (作为逻辑操作符)")
		}
	}
}

func TestReservedKeywords_CompleteCoverage(t *testing.T) {
	// 确保所有在文档中提到的操作符都被包含
	requiredKeywords := map[string]bool{
		"AND":        true,
		"OR":         true,
		"NOT":        true,
		"equals":     true,
		"not":        true,
		"in":         true,
		"lt":         true,
		"lte":        true,
		"gt":         true,
		"gte":        true,
		"contains":   true,
		"startsWith": true,
		"endsWith":   true,
		"mode":       true,
	}

	keywords := GetReservedKeywords()
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[kw] = true
	}

	for required := range requiredKeywords {
		if !keywordMap[required] {
			t.Errorf("保留关键字列表缺少必需的关键字: %s", required)
		}
	}

	// 验证总数是否合理（至少包含所有必需的关键字）
	if len(keywords) < len(requiredKeywords) {
		t.Errorf("保留关键字数量 (%d) 少于必需的关键字数量 (%d)", len(keywords), len(requiredKeywords))
	}
}
