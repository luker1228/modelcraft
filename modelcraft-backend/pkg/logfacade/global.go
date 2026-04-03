package logfacade

import (
	"context"
)

// 全局日志实例 - 仅供服务初始化使用
var defaultLogger Logger

// init initializes the default logger at package load time.
// This ensures a logger is always available even before main() runs.
func init() {
	// 创建一个基本的默认日志实例
	config := Config{
		Level:      InfoLevel,
		OutputPath: "stdout",
	}
	logger, err := NewWithSkipStack(config, 2)
	if err != nil {
		// If initialization fails, panic as logging is essential
		panic("failed to initialize default logger: " + err.Error())
	}
	defaultLogger = logger
}

// SetDefault 设置全局默认日志实例。
// 通常由 Setup 方法内部调用，开发者不应直接调用此方法。
//
// 注意：此方法仅在服务启动时使用，业务代码禁止调用此方法。
//
// 参数：
//   - logger: 要设置为全局默认的 Logger 实例
//
// 示例（不推荐）：
//
//	// 除非有特殊需求，否则应该使用 Setup 方法
//	customLogger, _ := New(config)
//	SetDefault(customLogger)
func SetDefault(logger Logger) {
	defaultLogger = logger
}

// getDefault 获取全局默认日志实例（私有方法）。
// 该方法仅供 logfacade 包内部使用。
//
// 返回：
//   - Logger: 全局默认的日志实例
func getDefault() Logger {
	return defaultLogger
}

// WithLogger 便捷设置方法
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}

// GetLogger 从上下文中获取 Logger 实例。
// 如果上下文中没有 Logger，则返回全局默认 Logger。
//
// 推荐业务代码使用此方法获取 Logger，以支持依赖注入和上下文传递。
func GetLogger(ctx context.Context) Logger {
	if logger, ok := ctx.Value(LoggerKey).(Logger); ok {
		return logger
	}
	return getDefault()
}
