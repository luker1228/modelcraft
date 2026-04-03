package query

import (
	"reflect"
	"testing"
)

// ============================================================================
// 常量测试
// ============================================================================

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FieldEquals", FieldEquals, "equals"},
		{"FieldNot", FieldNot, "not"},
		{"FieldIn", FieldIn, "in"},
		{"FieldContains", FieldContains, "contains"},
		{"FieldStartsWith", FieldStartsWith, "startsWith"},
		{"FieldEndsWith", FieldEndsWith, "endsWith"},
		{"FieldMode", FieldMode, "mode"},
		{"FieldLt", FieldLt, "lt"},
		{"FieldLte", FieldLte, "lte"},
		{"FieldGt", FieldGt, "gt"},
		{"FieldGte", FieldGte, "gte"},
		{"QueryModeDefault", QueryModeDefault, "default"},
		{"QueryModeInsensitive", QueryModeInsensitive, "insensitive"},
		{"QueryModeName", QueryModeName, "QueryMode"},
		{"LogicalOperatorAND", LogicalOperatorAND, "AND"},
		{"LogicalOperatorOR", LogicalOperatorOR, "OR"},
		{"LogicalOperatorNOT", LogicalOperatorNOT, "NOT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("常量 %s = %v, 期望 %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// ============================================================================
// 操作符检查函数测试
// ============================================================================

func TestIsLogicalOperator(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected bool
	}{
		{"AND操作符", LogicalOperatorAND, true},
		{"OR操作符", LogicalOperatorOR, true},
		{"非逻辑操作符_equals", FieldEquals, false},
		{"非逻辑操作符_gt", FieldGt, false},
		{"空字符串", "", false},
		{"随机字符串", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLogicalOperator(tt.operator)
			if result != tt.expected {
				t.Errorf("IsLogicalOperator(%s) = %v, 期望 %v", tt.operator, result, tt.expected)
			}
		})
	}
}

func TestIsComparisonOperator(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected bool
	}{
		{"equals操作符", FieldEquals, true},
		{"not操作符", FieldNot, true},
		{"gt操作符", FieldGt, true},
		{"gte操作符", FieldGte, true},
		{"lt操作符", FieldLt, true},
		{"lte操作符", FieldLte, true},
		{"contains操作符", FieldContains, true},
		{"startsWith操作符", FieldStartsWith, true},
		{"endsWith操作符", FieldEndsWith, true},
		{"in操作符", FieldIn, true},
		{"非比较操作符_AND", LogicalOperatorAND, false},
		{"非比较操作符_OR", LogicalOperatorOR, false},
		{"空字符串", "", false},
		{"随机字符串", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsComparisonOperator(tt.operator)
			if result != tt.expected {
				t.Errorf("IsComparisonOperator(%s) = %v, 期望 %v", tt.operator, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Condition 类型方法测试
// ============================================================================

func TestCondition_And(t *testing.T) {
	cond1 := Condition{"name": Eq("john")}
	cond2 := Condition{"age": Gt(18)}
	cond3 := Condition{"status": Eq("active")}

	result := cond1.And(cond2, cond3)

	expected := Condition{
		LogicalOperatorAND: []Condition{cond1, cond2, cond3},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Condition.And() = %v, 期望 %v", result, expected)
	}
}

func TestCondition_Or(t *testing.T) {
	cond1 := Condition{"role": Eq("admin")}
	cond2 := Condition{"role": Eq("moderator")}

	result := cond1.Or(cond2)

	expected := Condition{
		LogicalOperatorOR: []Condition{cond1, cond2},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Condition.Or() = %v, 期望 %v", result, expected)
	}
}

func TestCondition_Not(t *testing.T) {
	cond := Condition{"status": Eq("deleted")}
	result := cond.Not()

	expected := Condition{
		LogicalOperatorNOT: cond,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Condition.Not() = %v, 期望 %v", result, expected)
	}
}

func TestCondition_ToMap(t *testing.T) {
	cond := Condition{"name": "john", "age": 25}
	result := cond.ToMap()

	expected := map[string]any{"name": "john", "age": 25}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Condition.ToMap() = %v, 期望 %v", result, expected)
	}
}

func TestCondition_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		expected  bool
	}{
		{"空条件", Condition{}, true},
		{"非空条件", Condition{"name": "john"}, false},
		{"nil条件", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.IsEmpty()
			if result != tt.expected {
				t.Errorf("Condition.IsEmpty() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestCondition_Clone(t *testing.T) {
	original := Condition{"name": "john", "age": 25}
	cloned := original.Clone()

	// 检查内容是否相同
	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Clone() 内容不匹配: original = %v, cloned = %v", original, cloned)
	}

	// 检查是否是不同的实例
	cloned["new_field"] = "new_value"
	if _, exists := original["new_field"]; exists {
		t.Error("Clone() 没有创建独立的副本")
	}
}

func TestCondition_Merge(t *testing.T) {
	cond1 := Condition{"name": "john", "age": 25}
	cond2 := Condition{"age": 30, "status": "active"}

	result := cond1.Merge(cond2)

	expected := Condition{"name": "john", "age": 30, "status": "active"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Condition.Merge() = %v, 期望 %v", result, expected)
	}

	// 检查是否修改了原始条件
	if !reflect.DeepEqual(cond1, expected) {
		t.Error("Merge() 应该修改原始条件")
	}
}

// ============================================================================
// 构造函数测试
// ============================================================================

func TestNewCondition(t *testing.T) {
	result := NewCondition()

	if result == nil {
		t.Error("NewCondition() 返回 nil")
	}

	if len(result) != 0 {
		t.Errorf("NewCondition() 应该返回空条件, 得到 %v", result)
	}
}

func TestNewConditionFrom(t *testing.T) {
	input := map[string]any{"name": "john", "age": 25}
	result := NewConditionFrom(input)

	expected := Condition{"name": "john", "age": 25}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewConditionFrom() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// FieldBuilder 测试
// ============================================================================

func TestField(t *testing.T) {
	builder := Field("name")

	if builder == nil {
		t.Error("Field() 返回 nil")
	}

	if builder.fieldName != "name" {
		t.Errorf("Field() fieldName = %s, 期望 'name'", builder.fieldName)
	}
}

func TestFieldBuilder_Eq(t *testing.T) {
	result := Field("name").Eq("john")
	expected := Condition{"name": Eq("john")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Eq() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Not(t *testing.T) {
	result := Field("status").Not("deleted")
	expected := Condition{"status": Not("deleted")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Not() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_In(t *testing.T) {
	result := Field("status").In("active", "pending")
	expected := Condition{"status": In("active", "pending")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.In() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Gt(t *testing.T) {
	result := Field("age").Gt(18)
	expected := Condition{"age": Gt(18)}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Gt() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Gte(t *testing.T) {
	result := Field("age").Gte(18)
	expected := Condition{"age": Gte(18)}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Gte() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Lt(t *testing.T) {
	result := Field("age").Lt(65)
	expected := Condition{"age": Lt(65)}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Lt() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Lte(t *testing.T) {
	result := Field("age").Lte(65)
	expected := Condition{"age": Lte(65)}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Lte() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Contains(t *testing.T) {
	result := Field("name").Contains("john")
	expected := Condition{"name": Contains("john")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Contains() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_StartsWith(t *testing.T) {
	result := Field("name").StartsWith("john")
	expected := Condition{"name": StartsWith("john")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.StartsWith() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_EndsWith(t *testing.T) {
	result := Field("name").EndsWith("doe")
	expected := Condition{"name": EndsWith("doe")}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.EndsWith() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_Between(t *testing.T) {
	result := Field("age").Between(18, 65)
	expected := Condition{"age": Between(18, 65)}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.Between() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_IsNull(t *testing.T) {
	result := Field("deleted_at").IsNull()
	expected := Condition{"deleted_at": IsNull()}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.IsNull() = %v, 期望 %v", result, expected)
	}
}

func TestFieldBuilder_IsNotNull(t *testing.T) {
	result := Field("created_at").IsNotNull()
	expected := Condition{"created_at": IsNotNull()}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FieldBuilder.IsNotNull() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// 便捷条件生成函数测试
// ============================================================================

func TestEq(t *testing.T) {
	result := Eq("john")
	expected := Condition{FieldEquals: "john"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Eq() = %v, 期望 %v", result, expected)
	}
}

func TestNot(t *testing.T) {
	result := Not("deleted")
	expected := Condition{FieldNot: "deleted"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Not() = %v, 期望 %v", result, expected)
	}
}

func TestIn(t *testing.T) {
	result := In("active", "pending", "completed")
	expected := Condition{FieldIn: []any{"active", "pending", "completed"}}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("In() = %v, 期望 %v", result, expected)
	}
}

func TestGt(t *testing.T) {
	result := Gt(18)
	expected := Condition{FieldGt: 18}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Gt() = %v, 期望 %v", result, expected)
	}
}

func TestGte(t *testing.T) {
	result := Gte(18)
	expected := Condition{FieldGte: 18}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Gte() = %v, 期望 %v", result, expected)
	}
}

func TestLt(t *testing.T) {
	result := Lt(65)
	expected := Condition{FieldLt: 65}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Lt() = %v, 期望 %v", result, expected)
	}
}

func TestLte(t *testing.T) {
	result := Lte(65)
	expected := Condition{FieldLte: 65}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Lte() = %v, 期望 %v", result, expected)
	}
}

func TestContains(t *testing.T) {
	result := Contains("john")
	expected := Condition{FieldContains: "john"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Contains() = %v, 期望 %v", result, expected)
	}
}

func TestStartsWith(t *testing.T) {
	result := StartsWith("john")
	expected := Condition{FieldStartsWith: "john"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("StartsWith() = %v, 期望 %v", result, expected)
	}
}

func TestEndsWith(t *testing.T) {
	result := EndsWith("doe")
	expected := Condition{FieldEndsWith: "doe"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("EndsWith() = %v, 期望 %v", result, expected)
	}
}

func TestContainsInsensitive(t *testing.T) {
	result := ContainsInsensitive("JOHN")
	expected := Condition{
		FieldContains: "JOHN",
		FieldMode:     QueryModeInsensitive,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ContainsInsensitive() = %v, 期望 %v", result, expected)
	}
}

func TestStartsWithInsensitive(t *testing.T) {
	result := StartsWithInsensitive("JOHN")
	expected := Condition{
		FieldStartsWith: "JOHN",
		FieldMode:       QueryModeInsensitive,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("StartsWithInsensitive() = %v, 期望 %v", result, expected)
	}
}

func TestEndsWithInsensitive(t *testing.T) {
	result := EndsWithInsensitive("DOE")
	expected := Condition{
		FieldEndsWith: "DOE",
		FieldMode:     QueryModeInsensitive,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("EndsWithInsensitive() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// 逻辑操作符测试
// ============================================================================

func TestAnd(t *testing.T) {
	cond1 := Eq("john")
	cond2 := Gt(18)

	result := And(cond1, cond2)
	expected := Condition{LogicalOperatorAND: []Condition{cond1, cond2}}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("And() = %v, 期望 %v", result, expected)
	}
}

func TestOr(t *testing.T) {
	cond1 := Eq("admin")
	cond2 := Eq("moderator")

	result := Or(cond1, cond2)
	expected := Condition{LogicalOperatorOR: []Condition{cond1, cond2}}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Or() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// 范围查询测试
// ============================================================================

func TestBetween(t *testing.T) {
	result := Between(18, 65)
	expected := And(Gte(18), Lte(65))

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Between() = %v, 期望 %v", result, expected)
	}
}

func TestNotBetween(t *testing.T) {
	result := NotBetween(18, 65)
	expected := Or(Lt(18), Gt(65))

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NotBetween() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// 空值检查测试
// ============================================================================

func TestIsNull(t *testing.T) {
	result := IsNull()
	expected := Eq(nil)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("IsNull() = %v, 期望 %v", result, expected)
	}
}

func TestIsNotNull(t *testing.T) {
	result := IsNotNull()
	expected := Not(nil)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("IsNotNull() = %v, 期望 %v", result, expected)
	}
}

// ============================================================================
// 集成测试 - 复杂场景
// ============================================================================

func TestComplexConditions(t *testing.T) {
	// 测试复杂的条件组合
	condition := And(
		Field("name").Contains("john"),
		Or(
			Field("age").Between(18, 65),
			Field("role").Eq("admin"),
		),
		LogicalNot(Field("status").In("deleted", "banned")),
	)

	// 验证结构是否正确
	if condition.IsEmpty() {
		t.Error("复杂条件不应该为空")
	}

	// 验证可以转换为map
	condMap := condition.ToMap()
	if condMap == nil {
		t.Error("条件应该能够转换为map")
	}
}

func TestChainedOperations(t *testing.T) {
	// 测试链式操作
	base := Field("name").Eq("john")
	result := base.And(
		Field("age").Gte(18),
	).Or(
		Field("role").Eq("admin"),
	)

	if result.IsEmpty() {
		t.Error("链式操作结果不应该为空")
	}
}

func TestConditionCloneIndependence(t *testing.T) {
	// 测试克隆的独立性
	original := Field("name").Eq("john")
	cloned := original.Clone()

	// 修改克隆不应该影响原始条件
	cloned.Merge(Field("age").Gt(18))

	if reflect.DeepEqual(original, cloned) {
		t.Error("克隆应该是独立的")
	}
}

// ============================================================================
// 基准测试
// ============================================================================

func BenchmarkEq(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Eq("test_value")
	}
}

func BenchmarkFieldBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Field("name").Eq("test_value")
	}
}

func BenchmarkComplexCondition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		And(
			Field("name").Contains("john"),
			Or(
				Field("age").Between(18, 65),
				Field("role").Eq("admin"),
			),
		)
	}
}

// ============================================================================
// NOT 逻辑操作符测试
// ============================================================================

func TestLogicalNot(t *testing.T) {
	tests := []struct {
		name     string
		input    Condition
		expected Condition
	}{
		{
			name:  "简单NOT条件",
			input: Field("status").Eq("draft"),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					"status": Condition{FieldEquals: "draft"},
				},
			},
		},
		{
			name:  "NOT包含条件",
			input: Field("title").Contains("SQL"),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					"title": Condition{FieldContains: "SQL"},
				},
			},
		},
		{
			name:  "NOT替代notIn语义",
			input: Field("status").In("draft", "archived"),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					"status": Condition{FieldIn: []any{"draft", "archived"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LogicalNot(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("LogicalNot() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestLogicalNotWithComplexConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    Condition
		expected Condition
	}{
		{
			name: "NOT AND组合",
			input: And(
				Field("published").Eq(true),
				Field("views").Gt(1000),
			),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					LogicalOperatorAND: []Condition{
						{"published": Condition{FieldEquals: true}},
						{"views": Condition{FieldGt: 1000}},
					},
				},
			},
		},
		{
			name: "NOT OR组合",
			input: Or(
				Field("status").Eq("draft"),
				Field("status").Eq("archived"),
			),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					LogicalOperatorOR: []Condition{
						{"status": Condition{FieldEquals: "draft"}},
						{"status": Condition{FieldEquals: "archived"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LogicalNot(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("LogicalNot() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestConditionNotMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    Condition
		expected Condition
	}{
		{
			name:  "Condition.Not()方法",
			input: Field("status").Eq("archived"),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					"status": Condition{FieldEquals: "archived"},
				},
			},
		},
		{
			name:  "链式调用.Not()",
			input: Field("title").Contains("draft"),
			expected: Condition{
				LogicalOperatorNOT: Condition{
					"title": Condition{FieldContains: "draft"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Not()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Condition.Not() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestMultipleNOTConditions(t *testing.T) {
	result := And(
		LogicalNot(Field("status").Eq("archived")),
		LogicalNot(Field("title").Contains("draft")),
	)

	expected := Condition{
		LogicalOperatorAND: []Condition{
			{
				LogicalOperatorNOT: Condition{
					"status": Condition{FieldEquals: "archived"},
				},
			},
			{
				LogicalOperatorNOT: Condition{
					"title": Condition{FieldContains: "draft"},
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Multiple NOT in AND = %v, 期望 %v", result, expected)
	}
}

func TestIsLogicalOperatorWithNOT(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		expected bool
	}{
		{"NOT操作符", LogicalOperatorNOT, true},
		{"NOT操作符大写", "NOT", true},
		{"AND操作符", LogicalOperatorAND, true},
		{"OR操作符", LogicalOperatorOR, true},
		{"非逻辑操作符_equals", FieldEquals, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLogicalOperator(tt.operator)
			if result != tt.expected {
				t.Errorf("IsLogicalOperator(%s) = %v, 期望 %v", tt.operator, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// 基准测试
// ============================================================================

func BenchmarkConditionClone(b *testing.B) {
	condition := And(
		Field("name").Eq("john"),
		Field("age").Gt(18),
		Field("status").In("active", "pending"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		condition.Clone()
	}
}
