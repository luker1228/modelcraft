package logfacade

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Logger 定义了日志记录接口，隐藏底层实现细节，支持多种日志库的无缝切换。
// 该接口提供了结构化日志记录的完整功能，包括基础日志方法、链式调用。
// 上下文通过每个方法的第一个参数传入，自动提取 request_id 等字段。
//
// 使用方式：
//
//	// 结构化日志
//	logger.Info(ctx, "请求开始处理")
//	logger.Error(ctx, "处理失败", logfacade.Err(err), logfacade.Stack(err))
//
//	// 格式化日志
//	logger.Infof(ctx, "created project: slug=%s", slug)
//
//	// 附加固定字段
//	repoLogger := logger.With(logfacade.String("component", "model-repo"))
//	repoLogger.Info(ctx, "fetching model", logfacade.String("id", id))
type Logger interface {
	// Debug records a debug-level log message with context.
	Debug(ctx context.Context, msg string, fields ...Field)

	// Info records an info-level log message with context.
	Info(ctx context.Context, msg string, fields ...Field)

	// Warn records a warn-level log message with context.
	Warn(ctx context.Context, msg string, fields ...Field)

	// Error records an error-level log message with context.
	Error(ctx context.Context, msg string, fields ...Field)

	// Fatal records a fatal-level log message with context and exits.
	Fatal(ctx context.Context, msg string, fields ...Field)

	// Debugf records a debug-level log message using printf-style formatting.
	Debugf(ctx context.Context, format string, args ...any)

	// Infof records an info-level log message using printf-style formatting.
	Infof(ctx context.Context, format string, args ...any)

	// Warnf records a warn-level log message using printf-style formatting.
	Warnf(ctx context.Context, format string, args ...any)

	// Errorf records an error-level log message using printf-style formatting.
	Errorf(ctx context.Context, format string, args ...any)

	// Fatalf records a fatal-level log message using printf-style formatting and exits.
	Fatalf(ctx context.Context, format string, args ...any)

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
//	logger.Info("用户登录", logfacade.String("username", "john"))
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
//	logger.Info("用户信息", logfacade.Int("user_id", 123))
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
//	logger.Info("统计信息", logfacade.Int64("total_count", 9223372036854775807))
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
//	logger.Info("性能指标", logfacade.Float64("response_time", 0.123))
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
//	logger.Info("系统状态", logfacade.Bool("is_running", true))
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
//	logger.Info("请求耗时", logfacade.Duration("elapsed", 150*time.Millisecond))
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
//	logger.Error("操作失败", logfacade.Err(err))
func Err(err error) Field {
	// 如果有堆栈信息
	if _, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		return Field{Key: "error", Value: fmt.Sprintf("%+v", err)}
	}

	// 普通错误
	return Field{Key: "error", Value: err}
}

// Stack 创建一个堆栈跟踪类型的日志字段。
// 该函数自动使用 "stack" 作为字段名，将错误的堆栈跟踪信息作为值。
// 用于打印"错误产生点的栈"，使用 fmt.Sprintf("%+v", err) 把 pkg/errors 的 stack 打出来。
//
// 参数：
//   - err: 错误对象（需要是 pkg/errors 包装的错误才有堆栈信息）
//
// 返回：
//   - Field: 包含键值对的日志字段，其中 Key 为 "stack"
//
// 示例：
//
//	// 带有堆栈跟踪的错误
//	err := pkgerrors.Wrap(originalErr, "context")
//	logger.Error("operation failed", logfacade.Stack(err))
//
//	// 将同时打印错误信息和堆栈跟踪
//	logger.Error("operation failed", logfacade.Err(err), logfacade.Stack(err))
func Stack(err error) Field {
	if err == nil {
		return Field{Key: "stack", Value: nil}
	}

	// 检查错误是否实现了 StackTrace 接口（pkg/errors 提供的接口）
	if _, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		// 使用 %+v 格式化，打印完整的堆栈跟踪信息
		return Field{Key: "stack", Value: fmt.Sprintf("%+v", err)}
	}

	// 如果没有堆栈信息，直接返回错误本身
	return Field{Key: "stack", Value: err}
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
//	logger.Info("复杂数据", logfacade.Any("user", user))
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
