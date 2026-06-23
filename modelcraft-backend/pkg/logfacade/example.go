package logfacade

// 这是一个 logfacade 日志接口的使用示例文档。
//
// 日志接口统一采用 printf 风格，签名固定为：
//
//	Debugf(ctx, format, args...)
//	Infof(ctx, format, args...)
//	Warnf(ctx, format, args...)
//	Errorf(ctx, err, format, args...)   // err 第二个参数，自动注入 error + stack 结构化字段
//	Fatalf(ctx, err, format, args...)
//
// 结构化额外字段通过 With 预附加；err 与堆栈由 Errorf 自动处理，调用方无需关心 %v 还是 %+v。
//
// ------------------------------------------------------------------
// 1. 初始化
// ------------------------------------------------------------------
//
//	logger, err := logfacade.New(logfacade.Config{
//	    Level:      logfacade.InfoLevel,
//	    OutputPath: "stdout", // 或文件路径
//	})
//	if err != nil {
//	    panic(err)
//	}
//	defer logger.Sync()
//
// ------------------------------------------------------------------
// 2. 基础日志（无 error）
// ------------------------------------------------------------------
//
//	logger.Debugf(ctx, "query sql: %s", sql)
//	logger.Infof(ctx, "project created: slug=%s", slug)
//	logger.Warnf(ctx, "slow query: rows=%d elapsed=%v", rows, elapsed)
//	logger.Errorf(ctx, nil, "just a message: %d", code) // err=nil，不注入堆栈
//
// ------------------------------------------------------------------
// 3. 错误日志（自动注入 error + stack）
// ------------------------------------------------------------------
//
// 错误通过 Errorf 的第二个参数传入。Logger 内部自动：
//   - 注入 "error" 结构化字段（err.Error()）
//   - 若 err 携带 stack trace（pkg/errors 等实现 StackTrace()），注入 "stack" 字段
//
// 调用方完全不关心格式串里是 %v 还是 %+v，堆栈都会出现：
//
//	if err != nil {
//	    logger.Errorf(ctx, err, "create project failed: slug=%s", slug)
//	    return err
//	}
//
// ------------------------------------------------------------------
// 4. 包装错误堆栈
// ------------------------------------------------------------------
//
// 要让 err 携带 stack trace，用 pkg/errors 包构造：
//
//	import pkgerrors "github.com/pkg/errors"
//
//	func createProject(name string) error {
//	    if err := validate(name); err != nil {
//	        return pkgerrors.Wrap(err, "createProject") // 自动带堆栈
//	    }
//	    return nil
//	}
//
// 然后调用方就能拿到完整堆栈：
//
//	if err := createProject("foo"); err != nil {
//	    logger.Errorf(ctx, err, "createProject failed") // 自动输出 stack 字段
//	}
//
// ------------------------------------------------------------------
// 5. 附加结构化字段（With）
// ------------------------------------------------------------------
//
// 业务字段通过 With 预附加，自动出现在每条日志：
//
//	repoLogger := logger.With(
//	    logfacade.String("component", "model-repo"),
//	    logfacade.String("org", orgName),
//	)
//
//	repoLogger.Infof(ctx, "fetching model: id=%s", id)
//	// -> log: msg="fetching model: id=xxx" component="model-repo" org="demo"
//
// 错误场景下，With 的字段会与 Errorf 注入的 error/stack 字段共存：
//
//	repoLogger.Errorf(ctx, err, "fetch failed: id=%s", id)
//	// -> log: msg="fetch failed: id=xxx" component="model-repo" org="demo" error="..." stack="..."
//
// ------------------------------------------------------------------
// 6. 上下文提取
// ------------------------------------------------------------------
//
// Logger 自动从 ctx 提取 request_id（若用 RequestIDKey 写入）：
//
//	ctx = context.WithValue(ctx, logfacade.RequestIDKey, "req-123")
//	logger.Infof(ctx, "开始处理")
//	// -> log: msg="开始处理" request_id="req-123"
//
// ------------------------------------------------------------------
// 7. Fatal（退出前自动输出堆栈并 flush）
// ------------------------------------------------------------------
//
//	if err := loadConfig(); err != nil {
//	    logger.Fatalf(ctx, err, "config load failed: path=%s", cfgPath)
//	    // 执行后进程自动 exit(1)
//	}
//
// ------------------------------------------------------------------
// 8. 典型业务调用样例
// ------------------------------------------------------------------
//
//	func (s *Service) CreateUser(ctx context.Context, name string) error {
//	    if err := s.repo.Create(ctx, name); err != nil {
//	        s.logger.Errorf(ctx, err, "create user failed: name=%s", name)
//	        return pkgerrors.Wrap(err, "CreateUser")
//	    }
//	    s.logger.Infof(ctx, "user created: name=%s", name)
//	    return nil
//	}
//
// ------------------------------------------------------------------
// 9. Errorf 使用方式
// ------------------------------------------------------------------
//
//	logger.Errorf(ctx, err, "msg")
//	// 自动注入 error 结构化字段
