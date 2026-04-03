package logfacade

// New 创建一个新的日志记录器实例。
// 这是业务代码应该使用的主要方法来创建 Logger 实例。
// 该方法会根据配置自动选择合适的日志实现（当前为 zap 实现）。
//
// 参数：
//   - config: 日志配置，包括级别、输出路径等
//
// 返回：
//   - Logger: 日志记录器接口实现
//   - error: 初始化过程中的错误（如文件打开失败等）
//
// 示例：
//
//	logger, err := New(Config{
//	    Level:      InfoLevel,
//	    OutputPath: "stdout",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer logger.Sync()
//	logger.Info("应用启动", String("version", "1.0.0"))
//
// 注意：
//   - 每个服务或组件通常应该创建自己的 Logger 实例
//   - 在程序退出前应调用 Sync() 方法确保所有日志已写入
func New(config Config) (Logger, error) {
	return newZapLogger(config, 1)
}

// NewWithSkipStack 创建一个新的日志记录器实例，并指定调用栈跳过的层数。
// 这个方法主要用于封装日志库或创建日志工厂函数时，
// 调整 caller 信息以显示正确的代码位置。
//
// 参数：
//   - config: 日志配置
//   - skipStack: 调用栈跳过的层数。通常 New() 使用 1，在包装函数中使用 2 或更多
//
// 返回：
//   - Logger: 日志记录器接口实现
//   - error: 初始化过程中的错误
//
// 示例：
//
//	// 在日志包装函数中使用，跳过包装层
//	func createLogger(config Config) (Logger, error) {
//	    return NewWithSkipStack(config, 2)  // 跳过本函数，显示调用者位置
//	}
//
// 注意：
//   - skipStack 值过大可能导致 caller 信息不准确
//   - 通常只有在实现日志工厂或中间件时才需要使用此方法
func NewWithSkipStack(config Config, skipStack int) (Logger, error) {
	return newZapLogger(config, skipStack)
}
