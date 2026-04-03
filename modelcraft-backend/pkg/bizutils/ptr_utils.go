package bizutils

// IntPtr 返回给定整数值的指针。
// 常用于初始化结构体中的指针字段或传递可选的整数参数。
func IntPtr(i int) *int {
	return &i
}

// StringPtr 返回给定字符串值的指针。
// 常用于初始化结构体中的指针字段或传递可选的字符串参数。
func StringPtr(s string) *string {
	return &s
}

// SafeString 返回给定字符串指针的值，如果指针为nil，则返回空字符串。
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BoolPtr 返回给定布尔值的指针。
// 常用于初始化结构体中的指针字段或传递可选的布尔参数。
func BoolPtr(b bool) *bool {
	return &b
}
