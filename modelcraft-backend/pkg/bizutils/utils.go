package bizutils

import (
	"context"
	"fmt"
	"modelcraft/pkg/logfacade"
	"runtime"
	"time"

	"github.com/sourcegraph/conc/panics"
)

// GoWithCtx 启动一个新的协程，支持上下文传递，并使用 conc/panics 捕获 panic
// 从上下文中获取 logger 记录 panic 信息
func GoWithCtx(ctx context.Context, fn func(context.Context)) {
	go func() {
		recovered := panics.Try(func() {
			fn(ctx)
		})

		if recovered != nil {
			logger := logfacade.GetLogger(ctx)

			logger.With(
				logfacade.Any("panic_value", recovered.Value),
				logfacade.String(logfacade.StackFieldKey, string(recovered.Stack)),
				logfacade.Int(logfacade.GoroutineCountKey, runtime.NumGoroutine()),
				logfacade.String(logfacade.PanicTimeKey, time.Now().Format(time.RFC3339Nano)),
			).Errorf(ctx, nil, "协程发生 panic")
		}
	}()
}

// ObservedTask 包装一个可能 panic 的任务，返回一个安全的 func() error。
// 配合 pool.New().WithErrors() 使用，用于结构化并发场景：
//
//	p := pool.New().WithMaxGoroutines(10).WithErrors()
//	p.Go(ObservedTask("sync-user", syncUser))
//	p.Go(ObservedTask("sync-order", syncOrder))
//	if err := p.Wait(); err != nil { return err }
func ObservedTask(name string, fn func() error) func() error {
	return func() (err error) {
		recovered := panics.Try(func() {
			err = fn()
		})

		if recovered != nil {
			logfacade.GetLogger(context.Background()).With(
				logfacade.String("task", name),
				logfacade.Any("panic_value", recovered.Value),
				logfacade.String("stack_trace", string(recovered.Stack)),
			).Errorf(context.Background(), nil, "task panic: %s", name)
			return fmt.Errorf("task %s panic: %w", name, recovered.AsError())
		}

		return err
	}
}
