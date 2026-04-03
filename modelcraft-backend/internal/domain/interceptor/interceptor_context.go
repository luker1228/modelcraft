package interceptor

// InterceptorContext holds all context information available to JavaScript interceptors
// during execution. This includes user identity, resource metadata, environment data,
// and the operation input.
type InterceptorContext struct {
	User        UserContext        `json:"user"`
	Resource    ResourceContext    `json:"resource"`
	Environment EnvironmentContext `json:"environment"`
	Input       interface{}        `json:"input"` // The operation input (FindManyInput, CreateOneInput, etc.)
}

// UserContext contains user identity and permission information
type UserContext struct {
	ID          string                 `json:"id"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"` // Additional user attributes (tenantId, department, etc.)
}

// ResourceContext describes the target model and operation
type ResourceContext struct {
	ModelID   uint64   `json:"modelId"`
	ModelName string   `json:"modelName"`
	Operation string   `json:"operation"` // findMany, createOne, etc.
	Fields    []string `json:"fields"`    // Fields being queried or modified
}

// EnvironmentContext provides request metadata
type EnvironmentContext struct {
	Timestamp string `json:"timestamp"` // ISO 8601 format
	RequestID string `json:"requestId"` // UUID v7 format
	IPAddress string `json:"ipAddress,omitempty"`
}

// NewInterceptorContext creates a new interceptor context
func NewInterceptorContext(
	user UserContext,
	resource ResourceContext,
	env EnvironmentContext,
	input interface{},
) *InterceptorContext {
	return &InterceptorContext{
		User:        user,
		Resource:    resource,
		Environment: env,
		Input:       input,
	}
}
