package logfacade

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger 是基于 go.uber.org/zap 库的 Logger 接口实现。
// 它提供高性能的结构化日志记录功能，支持多种输出方式和日志级别。
type ZapLogger struct {
	// logger 是底层的 zap 日志实例
	logger *zap.Logger
}

// newZapLogger 创建基于 zap 的日志实例。
// 该函数是内部工厂方法，根据配置初始化 zap Logger。
//
// 参数：
//   - config: 日志配置
//   - skipStack: 调用栈跳过层数，用于正确显示 caller 信息
//
// 返回：
//   - Logger: 日志记录器接口实现
//   - error: 初始化错误
func newZapLogger(config Config, skipStack int) (Logger, error) {
	var core zapcore.Core

	if config.OutputPath == "stdout" {
		core = zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(os.Stdout),
			getZapLevel(config.Level),
		)
	} else {
		// 使用 lumberjack 进行日志轮转
		writer := &lumberjack.Logger{
			Filename:   config.OutputPath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		core = zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(writer),
			getZapLevel(config.Level),
		)
	}

	// 添加调用者信息
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(skipStack))

	return &ZapLogger{logger: zapLogger}, nil
}

// Debugf records a debug-level log message using printf-style formatting.
func (z *ZapLogger) Debugf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Debug(msg, contextFields...)
}

// Infof records an info-level log message using printf-style formatting.
func (z *ZapLogger) Infof(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Info(msg, contextFields...)
}

// Warnf records a warn-level log message using printf-style formatting.
func (z *ZapLogger) Warnf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Warn(msg, contextFields...)
}

// Errorf records an error-level log message using printf-style formatting.
// When err is non-nil, it (and its stack trace if available) is recorded as
// structured fields, independent of the format string.
func (z *ZapLogger) Errorf(ctx context.Context, err error, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	fields := z.extractContextFields(ctx)
	fields = append(fields, z.extractErrFields(err)...)
	z.logger.Error(msg, fields...)
}

// Fatalf records a fatal-level log message using printf-style formatting and exits.
// Stack trace extraction behaves the same as Errorf.
func (z *ZapLogger) Fatalf(ctx context.Context, err error, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	fields := z.extractContextFields(ctx)
	fields = append(fields, z.extractErrFields(err)...)
	z.logger.Fatal(msg, fields...)
}

// extractErrFields produces structured fields from an error:
// always an "error" field; additionally a "stack" field when the error
// carries a stack trace (via StackTrace() []uintptr interface).
func (z *ZapLogger) extractErrFields(err error) []zap.Field {
	if err == nil {
		return nil
	}
	fields := []zap.Field{
		zap.String(ErrorFieldKey, err.Error()),
	}
	// walk error chain to find stack tracer
	for e := err; e != nil; e = errors.Unwrap(e) {
		if st, ok := e.(interface{ StackTrace() []uintptr }); ok {
			fields = append(fields, zap.String(StackFieldKey, fmt.Sprintf("%+v", st)))
			break
		}
	}
	return fields
}

// With creates a child Logger with the given fields pre-attached to all subsequent log calls.
//
// 示例：
//
//	serviceLogger := logger.With(String("service", "user-service"))
//	serviceLogger.Infof(ctx, "服务启动")  // 自动附加 service=user-service
func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: z.logger.With(z.convertFields(fields)...),
	}
}

// Sync 同步日志缓冲区，确保所有待输出的日志都被写入到目标位置。
// 应在程序退出前调用以确保日志不丢失。
func (z *ZapLogger) Sync() error {
	err := z.logger.Sync()
	// 忽略 stdout/stderr 的 sync 错误，这在某些环境中是正常的
	if err != nil {
		errStr := err.Error()
		// 检查是否是 stdout/stderr 的 sync 错误
		if errStr == "sync /dev/stdout: bad file descriptor" ||
			errStr == "sync /dev/stderr: bad file descriptor" ||
			errStr == "sync /dev/stdout: inappropriate ioctl for device" ||
			errStr == "sync /dev/stderr: inappropriate ioctl for device" {
			return nil // 忽略这些错误
		}
	}
	return err
}

// convertFields 将自定义 Field 切片转换为 zap 库的 Field 切片。
func (z *ZapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// extractContextFields 从上下文中提取日志字段，如 requestId 等。
func (z *ZapLogger) extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// 提取常见的上下文字段
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if rid, ok := requestID.(string); ok {
			fields = append(fields, zap.Any(RequestIDKey, rid))
		}
	}

	return fields
}

// getEncoder 获取 zap 编码器。
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

// getZapLevel 将自定义的 Level 转换为 zap 库的 zapcore.Level。
func getZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
