package logfacade

import (
	"context"
	"os"
	"testing"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// TestWriteToFile 测试将日志写入文件
func TestWriteToFile(t *testing.T) {
	// 创建临时日志目录
	logDir := "./test_logs"
	err := os.MkdirAll(logDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create logs directory: %v", err)
	}

	// 测试结束后清理
	defer func() {
		os.RemoveAll(logDir)
	}()

	// 指定具体的文件路径
	logFile := logDir + "/test.log"
	jsonConfig := Config{
		Level:      InfoLevel,
		OutputPath: logFile, // 指定具体文件路径
		MaxSize:    10,      // 10MB
		MaxBackups: 3,
		MaxAge:     7, // 7天
		Compress:   true,
	}

	jsonLogger, err := New(jsonConfig)
	if err != nil {
		t.Fatalf("Failed to create JSON logger: %v", err)
	}

	jsonLogger.Info(context.Background(), "JSON 格式日志",
		String("type", "json"),
		Int("count", 42),
		String("test", "write_to_file"),
	)

	// 确保日志写入
	err = jsonLogger.Sync()
	if err != nil {
		t.Logf("Sync warning: %v", err) // 使用 Logf 而不是 Fatalf，因为 sync 错误可能是正常的
	}

	// 验证文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %s", logFile)
	}

	// 读取文件内容验证
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("Log file is empty")
	}

	t.Logf("日志文件已创建: %s", logFile)
	t.Logf("日志内容: %s", string(content))
}

// TestWriteToStdout 测试不同的日志格式
func TestWriteToStdout(t *testing.T) {
	// JSON 格式
	jsonConfig := Config{
		Level:      InfoLevel,
		OutputPath: "./logs",
	}

	jsonLogger, err := New(jsonConfig)
	if err != nil {
		t.Fatalf("Failed to create JSON logfacade: %v", err)
	}

	jsonLogger.Info(context.Background(), "JSON 格式日志",
		String("type", "json"),
		Int("count", 42),
	)

	// Console 格式
	consoleConfig := Config{
		Level:      InfoLevel,
		OutputPath: "stdout",
	}

	consoleLogger, err := New(consoleConfig)
	if err != nil {
		t.Fatalf("Failed to create console logfacade: %v", err)
	}

	consoleLogger.Info(context.Background(), "Console 格式日志",
		String("type", "console"),
		Int("count", 42),
	)
}

// TestLogLevels 测试不同的日志级别
func TestLogLevels(t *testing.T) {
	config := Config{
		Level:      DebugLevel,
		OutputPath: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logfacade: %v", err)
	}

	logger.Debug(context.Background(), "这是调试信息")
	logger.Info(context.Background(), "这是信息日志")
	logger.Warn(context.Background(), "这是警告日志")
	logger.Error(context.Background(), "这是错误日志")
	// logfacade.Fatal("这是致命错误") // 会退出程序，测试时注释掉
}

// TestContextualLogging 测试上下文日志
func TestContextualLogging(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		OutputPath: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logfacade: %v", err)
	}

	// 模拟 HTTP 请求处理
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-789")
	ctx = context.WithValue(ctx, "user_id", 999)

	// 使用 ctx 进行日志
	logger.Info(ctx, "请求开始处理")
	logger.Info(ctx, "数据库查询", String("table", "users"))
	logger.Info(ctx, "请求处理完成", Int("status_code", 200))

	// 测试 With 和 ctx 的组合使用
	logger.With(String("module", "auth")).Info(ctx, "认证模块日志")
}

// TestErrorLoggingWithStack 测试错误日志与堆栈跟踪
func TestErrorLoggingWithStack(t *testing.T) {
	config := Config{
		Level:      ErrorLevel,
		OutputPath: "stdout",
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Example 1: Log with error message only (no stack)
	simpleErr := pkgerrors.New("simple error")
	logger.Error(context.Background(), "Operation failed", Err(simpleErr))

	// Example 2: Log with stack trace
	dbErr := pkgerrors.New("database connection timeout")
	wrappedErr := pkgerrors.Wrap(dbErr, "failed to initialize database pool")
	logger.Error(context.Background(), "Database initialization failed", Stack(wrappedErr))

	// Example 3: Log with both error message and stack trace
	// This is useful when you want error message in 'error' field
	// and full stack trace in 'stack' field
	apiErr := pkgerrors.New("API endpoint not found")
	contextErr := pkgerrors.Wrap(apiErr, "failed to process request")
	logger.Error(context.Background(), "Request processing failed",
		Err(contextErr),   // Error message (with stack if available)
		Stack(contextErr), // Full stack trace
		String("endpoint", "/api/v1/users"),
		Int("status", 404),
	)
}

// BenchmarkLogging 性能基准测试
func BenchmarkLogging(b *testing.B) {
	config := Config{
		Level:      InfoLevel,
		OutputPath: "/dev/null", // 避免 I/O 影响性能测试
	}

	logger, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create logfacade: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(context.Background(), "性能测试日志",
				String("operation", "benchmark"),
				Int("iteration", 1),
				Duration("duration", time.Microsecond),
			)
		}
	})
}
