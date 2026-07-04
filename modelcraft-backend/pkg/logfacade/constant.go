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
	StatusKey      string = "status"       // HTTP 状态码
	URLKey         string = "url"          // 完整 URL
	RemoteAddrKey  string = "remote_addr"  // 客户端地址
	UserAgentKey   string = "user_agent"   // User-Agent
	ContentTypeKey string = "content_type" // Content-Type

	// 请求/响应体
	RequestBodyKey  string = "request_body"  // 请求体
	ResponseBodyKey string = "response_body" // 响应体
	NetOutKey       string = "net_out"       // 响应大小（字节）

	// 数据库相关
	SQLKey       string = "sql"       // SQL 语句
	SQLArgsKey   string = "sql_args"  // SQL 参数
	RowsKey      string = "rows"      // 影响行数
	ElapsedKey   string = "elapsed"   // SQL/外部系统调用耗时
	ThresholdKey string = "threshold" // 阈值

	// 追踪相关
	ClientRequestIDKey string = "client_request_id" // 客户端请求 ID
	TraceIDKey         string = "trace_id"          // W3C Trace ID
	SpanIDKey          string = "span_id"           // W3C Span ID
	DurationKey        string = "duration"          // 耗时（ms）

	// 业务相关
	ProjectKey string = "project" // 项目标识
	ActionKey  string = "action"  // 接口动作（X-TC-Action）

	// Panic 相关
	PanicTimeKey      string = "panic_time"      // panic 发生时间
	GoroutineCountKey string = "goroutine_count" // 协程数量
)
