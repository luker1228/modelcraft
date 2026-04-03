// Package bizutils 提供业务逻辑相关的实用工具函数和接口。
// 包括时间处理、类型检查、JSON解析和指针操作等功能。
package bizutils

import "time"

// ParseTime 解析时间字符串为 time.Time 对象。
// 使用标准的 time.DateTime 格式进行解析。
// 如果解析失败，返回 Unix 时间戳 0（1970-01-01 00:00:00 UTC）。
func ParseTime(times string) time.Time {
	parsedTime, err := time.Parse(time.DateTime, times)
	if err != nil {
		return time.Unix(0, 0)
	}
	return parsedTime
}

// FormatTime 将 time.Time 对象格式化为标准字符串格式 "2006-01-02 15:04:05"。
// 为了确保输出在不同机器/时区下可预测，这里统一按 UTC 格式化。
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.DateTime)
}
