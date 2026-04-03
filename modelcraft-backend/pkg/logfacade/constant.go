package logfacade

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	LoggerKey contextKey = "loggerKey"
)

// Log field keys - 日志字段 key 常量
// 所有日志字段必须使用这些常量，禁止硬编码字符串
const (
	// 错误相关
	ErrorFieldKey string = "error" // 错误信息字段
	StackFieldKey string = "stack" // 堆栈跟踪字段

	// 请求相关
	RequestIDKey   string = "request_id"   // 请求 ID
	MethodKey      string = "method"       // HTTP 方法
	PathKey        string = "path"         // 请求路径
	URLKey         string = "url"          // 完整 URL
	RemoteAddrKey  string = "remote_addr"  // 客户端地址
	UserAgentKey   string = "user_agent"   // User-Agent
	StatusCodeKey  string = "status_code"  // HTTP 状态码
	ContentTypeKey string = "content_type" // Content-Type

	// 请求/响应体
	RequestBodyKey  string = "request_body"  // 请求体
	ResponseBodyKey string = "response_body" // 响应体
	BodyPreviewKey  string = "body_preview"  // 响应体预览
	BodyLengthKey   string = "body_length"   // 响应体长度
	SizeKey         string = "size"          // 响应大小（字节）

	// 数据库相关
	SQLKey       string = "sql"       // SQL 语句
	SQLArgsKey   string = "sql_args"  // SQL 参数
	RowsKey      string = "rows"      // 影响行数
	ElapsedKey   string = "elapsed"   // 耗时
	ThresholdKey string = "threshold" // 阈值

	// 业务相关
	ProjectKey string = "project" // 项目标识

	// 时间相关
	LatencyKey string = "latency" // 延迟/耗时

	// Panic 相关
	PanicTimeKey      string = "panic_time"      // panic 发生时间
	StackTraceKey     string = "stack_trace"     // 详细堆栈跟踪
	GoroutineCountKey string = "goroutine_count" // 协程数量
)
