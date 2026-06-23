package logfacade

import (
	"context"
	"time"
)

// Logger 定义了日志记录接口，隐藏底层实现细节，支持多种日志库的无缝切换。
// 该接口统一采用 printf 风格的日志记录方法，上下文通过第一个参数传入，
// 自动提取 request_id 等字段。
//
// 使用方式：
//
//	// 格式化日志
//	logger.Infof(ctx, "created project: slug=%s", slug)
//	logger.Errorf(ctx, "处理失败: %v", err)
//
//	// 附加固定字段（结构化字段通过 With 预附加）
//	repoLogger := logger.With(logfacade.String("component", "model-repo"))
//	repoLogger.Infof(ctx, "fetching model: id=%s", id)
type Logger interface {
	// Debugf records a debug-level log message using printf-style formatting.
	Debugf(ctx context.Context, format string, args ...any)

	// Infof records an info-level log message using printf-style formatting.
	Infof(ctx context.Context, format string, args ...any)

	// Warnf records a warn-level log message using printf-style formatting.
	Warnf(ctx context.Context, format string, args ...any)

	// Errorf records an error-level log message with printf-style formatting and
	// automatic stack trace extraction from the provided error.
	// The error is recorded as a structured "error" field; if the error carries
	// a stack trace (e.g. pkg/errors), it is also recorded in a structured "stack" field.
	// Pass nil when there is no associated error.
	Errorf(ctx context.Context, err error, format string, args ...any)

	// Fatalf records a fatal-level log message with printf-style formatting and exits.
	// Stack trace extraction behaves the same as Errorf.
	Fatalf(ctx context.Context, err error, format string, args ...any)

	// With creates a child Logger with the given fields pre-attached to all subsequent log calls.
	With(fields ...Field) Logger

	// Sync flushes any buffered log entries. Call before program exit.
	Sync() error
}

// Field 表示一个日志字段，由字段名和值组成。
// 日志字段用于为日志消息提供结构化的上下文信息。
type Field struct {
	// Key 是字段的名称
	Key string
	// Value 是字段的值，可以是任意类型
	Value interface{}
}

// Level 表示日志级别，用于控制日志的详细程度。
type Level string

const (
	// DebugLevel 调试级别，用于开发调试，包含最详细的信息
	DebugLevel Level = "debug"
	// InfoLevel 信息级别，用于记录常规的应用事件
	InfoLevel = "info"
	// WarnLevel 警告级别，用于记录潜在的问题
	WarnLevel = "warn"
	// ErrorLevel 错误级别，用于记录运行时错误
	ErrorLevel = "error"
	// FatalLevel 致命错误级别，用于不可恢复的错误，会导致程序退出
	FatalLevel = "fatal"
)

// Config 包含日志记录器的配置选项。
type Config struct {
	// Level 设置日志的最低输出级别，低于该级别的日志不会被输出
	Level Level `json:"level" mapstructure:"level"`
	// OutputPath 日志输出路径，可以是 "stdout" 或具体的文件路径
	OutputPath string `json:"output_path" mapstructure:"output_path"`
	// MaxSize 日志文件的最大大小，单位为 MB，超过此大小时日志文件会轮转
	MaxSize int `json:"max_size" mapstructure:"max_size"`
	// MaxBackups 保留的旧日志文件数量
	MaxBackups int `json:"max_backups" mapstructure:"max_backups"`
	// MaxAge 保留旧日志文件的最大天数
	MaxAge int `json:"max_age" mapstructure:"max_age"`
	// Compress 是否压缩被轮转的旧日志文件
	Compress bool `json:"compress" mapstructure:"compress"`
}

// String 创建一个字符串类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 字符串值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.String("username", "john")).Infof(ctx, "用户登录")
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

// Int 创建一个整数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 整数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Int("user_id", 123)).Infof(ctx, "用户信息")
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 创建一个 64 位整数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 64 位整数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Int64("total_count", 9223372036854775807)).Infof(ctx, "统计信息")
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Float64 创建一个浮点数类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 浮点数值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Float64("response_time", 0.123)).Infof(ctx, "性能指标")
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Bool 创建一个布尔类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 布尔值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Bool("is_running", true)).Infof(ctx, "系统状态")
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Duration 创建一个时间间隔类型的日志字段。
//
// 参数：
//   - key: 字段名称
//   - val: 时间间隔值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Duration("elapsed", 150*time.Millisecond)).Infof(ctx, "请求耗时")
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Err 创建一个错误类型的日志字段。
// 该函数自动使用 "error" 作为字段名，将错误信息作为值。
//
// 参数：
//   - err: 错误对象
//
// 返回：
//   - Field: 包含键值对的日志字段，其中 Key 为 "error"
//
// 示例：
//
//	logger.With(logfacade.Err(err)).Errorf(ctx, "操作失败")
func Err(err error) Field {
	return Field{Key: ErrorFieldKey, Value: err}
}



// Any 创建一个任意类型的日志字段。
// 该函数用于记录不属于上述任何预定义类型的值。
//
// 参数：
//   - key: 字段名称
//   - val: 任意类型的值
//
// 返回：
//   - Field: 包含键值对的日志字段
//
// 示例：
//
//	logger.With(logfacade.Any("user", user)).Infof(ctx, "复杂数据")
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
