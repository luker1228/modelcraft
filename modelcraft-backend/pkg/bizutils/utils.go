package bizutils

import (
	"context"
	"modelcraft/pkg/logfacade"
	"runtime"
	"time"
)

// GoWithCtx 启动一个新的协程，支持上下文传递，并处理可能的 panic
// 从上下文中获取 logger 记录 panic 信息
func GoWithCtx(ctx context.Context, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 从上下文中获取 logger
				logger := logfacade.GetLogger(ctx)

				// 获取详细的调用栈信息
				stack := make([]byte, 16384)         // 16KB 缓冲区
				length := runtime.Stack(stack, true) // 获取所有协程信息
				stackTrace := string(stack[:length])

				// 获取协程统计信息
				goroutineCount := runtime.NumGoroutine()
				panicTime := time.Now()

				// 记录详细的 panic 信息，包含上下文
				logger.Error(ctx, "协程发生 panic",
					logfacade.Any("panic_value", r),
					logfacade.String("stack_trace", stackTrace),
					logfacade.Int("goroutine_count", goroutineCount),
					logfacade.String("panic_time", panicTime.Format(time.RFC3339Nano)),
				)
			}
		}()

		// 执行用户函数，传递上下文
		fn(ctx)
	}()
}
