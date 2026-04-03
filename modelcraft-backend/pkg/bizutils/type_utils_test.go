package bizutils

import (
	"testing"
)

// TestIsArrayType_WithSlice 测试识别切片类型
func TestIsArrayType_WithSlice(t *testing.T) {
	slice := []int{1, 2, 3}
	result := IsArrayType(slice)

	if !result {
		t.Errorf("Expected true for slice, got %v", result)
	}
}

// TestIsArrayType_WithArray 测试识别数组类型
func TestIsArrayType_WithArray(t *testing.T) {
	array := [3]int{1, 2, 3}
	result := IsArrayType(array)

	if !result {
		t.Errorf("Expected true for array, got %v", result)
	}
}

// TestIsArrayType_WithStringSlice 测试识别字符串切片
func TestIsArrayType_WithStringSlice(t *testing.T) {
	stringSlice := []string{"a", "b", "c"}
	result := IsArrayType(stringSlice)

	if !result {
		t.Errorf("Expected true for string slice, got %v", result)
	}
}

// TestIsArrayType_WithEmptySlice 测试识别空切片
func TestIsArrayType_WithEmptySlice(t *testing.T) {
	emptySlice := []int{}
	result := IsArrayType(emptySlice)

	if !result {
		t.Errorf("Expected true for empty slice, got %v", result)
	}
}

// TestIsArrayType_WithInt 测试非数组类型（整数）
func TestIsArrayType_WithInt(t *testing.T) {
	value := 42
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for int, got %v", result)
	}
}

// TestIsArrayType_WithString 测试非数组类型（字符串）
func TestIsArrayType_WithString(t *testing.T) {
	value := "hello"
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for string, got %v", result)
	}
}

// TestIsArrayType_WithStruct 测试非数组类型（结构体）
func TestIsArrayType_WithStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	value := TestStruct{Name: "test", Age: 25}
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for struct, got %v", result)
	}
}

// TestIsArrayType_WithMap 测试非数组类型（map）
func TestIsArrayType_WithMap(t *testing.T) {
	value := map[string]int{"a": 1, "b": 2}
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for map, got %v", result)
	}
}

// TestIsArrayType_WithNil 测试 nil 值
func TestIsArrayType_WithNil(t *testing.T) {
	var value any = nil
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for nil, got %v", result)
	}
}

// TestIsArrayType_WithNilSlice 测试 nil 切片
func TestIsArrayType_WithNilSlice(t *testing.T) {
	var nilSlice []int
	result := IsArrayType(nilSlice)

	if !result {
		t.Errorf("Expected true for nil slice, got %v", result)
	}
}

// TestIsArrayType_WithInterfaceSlice 测试 any 切片
func TestIsArrayType_WithInterfaceSlice(t *testing.T) {
	slice := []any{1, "two", 3.0}
	result := IsArrayType(slice)

	if !result {
		t.Errorf("Expected true for any slice, got %v", result)
	}
}

// TestIsArrayType_WithBool 测试非数组类型（布尔）
func TestIsArrayType_WithBool(t *testing.T) {
	value := true
	result := IsArrayType(value)

	if result {
		t.Errorf("Expected false for bool, got %v", result)
	}
}
