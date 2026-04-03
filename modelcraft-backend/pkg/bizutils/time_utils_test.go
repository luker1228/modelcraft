package bizutils

import (
	"testing"
	"time"
)

// TestParseTime_ValidFormat 测试解析有效的标准时间格式
func TestParseTime_ValidFormat(t *testing.T) {
	input := "2023-12-25 10:30:45"
	result := ParseTime(input)

	expected := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestParseTime_InvalidFormat 测试解析无效格式，应返回 Unix 时间戳 0
func TestParseTime_InvalidFormat(t *testing.T) {
	input := "invalid-date-format"
	result := ParseTime(input)

	expected := time.Unix(0, 0)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestParseTime_EmptyString 测试解析空字符串
func TestParseTime_EmptyString(t *testing.T) {
	input := ""
	result := ParseTime(input)

	expected := time.Unix(0, 0)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestParseTime_EdgeCase_NewYear 测试解析年初时间
func TestParseTime_EdgeCase_NewYear(t *testing.T) {
	input := "2024-01-01 00:00:00"
	result := ParseTime(input)

	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestFormatTime_NormalTime 测试格式化正常时间
func TestFormatTime_NormalTime(t *testing.T) {
	input := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	result := FormatTime(input)

	expected := "2023-12-25 10:30:45"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestFormatTime_UnixEpoch 测试格式化 Unix 时间戳 0
func TestFormatTime_UnixEpoch(t *testing.T) {
	input := time.Unix(0, 0)
	result := FormatTime(input)

	expected := "1970-01-01 00:00:00"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestFormatTime_ZeroTime 测试格式化零值时间
func TestFormatTime_ZeroTime(t *testing.T) {
	input := time.Time{}
	result := FormatTime(input)

	expected := "0001-01-01 00:00:00"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestFormatTime_MidnightTime 测试格式化午夜时间
func TestFormatTime_MidnightTime(t *testing.T) {
	input := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	result := FormatTime(input)

	expected := "2023-12-31 23:59:59"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestParseTimeAndFormatTime_RoundTrip 测试解析和格式化的往返转换
func TestParseTimeAndFormatTime_RoundTrip(t *testing.T) {
	original := "2023-06-15 14:30:20"
	parsed := ParseTime(original)
	formatted := FormatTime(parsed)

	if formatted != original {
		t.Errorf("Round-trip failed: %s -> %v -> %s", original, parsed, formatted)
	}
}
