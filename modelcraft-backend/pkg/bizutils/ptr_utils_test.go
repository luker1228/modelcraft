package bizutils

import (
	"testing"
)

// TestIntPtr_ValidValue 测试创建正整数指针并验证值
func TestIntPtr_ValidValue(t *testing.T) {
	value := 42
	ptr := IntPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

// TestIntPtr_ZeroValue 测试零值指针
func TestIntPtr_ZeroValue(t *testing.T) {
	value := 0
	ptr := IntPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

// TestIntPtr_NegativeValue 测试负数值指针
func TestIntPtr_NegativeValue(t *testing.T) {
	value := -999
	ptr := IntPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

// TestIntPtr_LargeValue 测试大数值指针
func TestIntPtr_LargeValue(t *testing.T) {
	value := 2147483647 // Max int32
	ptr := IntPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %d, got %d", value, *ptr)
	}
}

// TestIntPtr_Independent 测试多个指针的独立性
func TestIntPtr_Independent(t *testing.T) {
	ptr1 := IntPtr(1)
	ptr2 := IntPtr(2)

	if ptr1 == ptr2 {
		t.Errorf("Expected different pointers, got same")
	}

	if *ptr1 != 1 || *ptr2 != 2 {
		t.Errorf("Pointers point to different values as expected")
	}
}

// TestStringPtr_ValidValue 测试创建字符串指针并验证值
func TestStringPtr_ValidValue(t *testing.T) {
	value := "hello world"
	ptr := StringPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %s, got %s", value, *ptr)
	}
}

// TestStringPtr_EmptyString 测试空字符串指针
func TestStringPtr_EmptyString(t *testing.T) {
	value := ""
	ptr := StringPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected empty string, got %s", *ptr)
	}
}

// TestStringPtr_LongString 测试长字符串指针
func TestStringPtr_LongString(t *testing.T) {
	value := "The quick brown fox jumps over the lazy dog. " +
		"This is a longer test string with multiple words and punctuation!"
	ptr := StringPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %s, got %s", value, *ptr)
	}
}

// TestStringPtr_SpecialCharacters 测试含特殊字符的字符串指针
func TestStringPtr_SpecialCharacters(t *testing.T) {
	value := "special!@#$%^&*()_+-=[]{}|;':\",./<>?\\`~"
	ptr := StringPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %s, got %s", value, *ptr)
	}
}

// TestStringPtr_UnicodeString 测试 Unicode 字符串指针
func TestStringPtr_UnicodeString(t *testing.T) {
	value := "你好世界 🌍 مرحبا العالم"
	ptr := StringPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected %s, got %s", value, *ptr)
	}
}

// TestBoolPtr_TrueValue 测试真值布尔指针
func TestBoolPtr_TrueValue(t *testing.T) {
	value := true
	ptr := BoolPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected true, got %v", *ptr)
	}
}

// TestBoolPtr_FalseValue 测试假值布尔指针
func TestBoolPtr_FalseValue(t *testing.T) {
	value := false
	ptr := BoolPtr(value)

	if ptr == nil {
		t.Errorf("Expected non-nil pointer, got nil")
		return
	}

	if *ptr != value {
		t.Errorf("Expected false, got %v", *ptr)
	}
}

// TestBoolPtr_Independent 测试多个布尔指针的独立性
func TestBoolPtr_Independent(t *testing.T) {
	ptrTrue := BoolPtr(true)
	ptrFalse := BoolPtr(false)

	if ptrTrue == ptrFalse {
		t.Errorf("Expected different pointers, got same")
	}

	if *ptrTrue != true || *ptrFalse != false {
		t.Errorf("Pointers have expected values")
	}
}

// TestPointerModification 测试通过指针修改值
func TestPointerModification(t *testing.T) {
	ptr := IntPtr(10)
	*ptr = 20

	if *ptr != 20 {
		t.Errorf("Expected modified value 20, got %d", *ptr)
	}
}

// TestStringPointerModification 测试通过字符串指针修改值
func TestStringPointerModification(t *testing.T) {
	ptr := StringPtr("original")
	*ptr = "modified"

	if *ptr != "modified" {
		t.Errorf("Expected modified string 'modified', got %s", *ptr)
	}
}
