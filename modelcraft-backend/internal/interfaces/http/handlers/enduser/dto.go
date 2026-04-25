package enduser

// ---------------------------------------------------------------------------
// Request/Response DTOs for End-User Management HTTP Handlers
// ---------------------------------------------------------------------------

// CreateEndUserRequest 创建终端用户请求
type CreateEndUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateEndUserResponse 创建终端用户响应
type CreateEndUserResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
	Error     *ErrorJSON   `json:"error,omitempty"`
}

// ListEndUsersRequest 列表查询请求参数（通过 query string 传递）
type ListEndUsersRequest struct {
	Search string `json:"search,omitempty"`
	First  int    `json:"first,omitempty"`
	After  string `json:"after,omitempty"`
}

// ListEndUsersResponse 列表查询响应
type ListEndUsersResponse struct {
	RequestID  string         `json:"requestId"`
	Items      []*EndUserJSON `json:"items"`
	TotalCount int64          `json:"totalCount"`
	PageInfo   *PageInfoJSON  `json:"pageInfo"`
	Error      *ErrorJSON     `json:"error,omitempty"`
}

// DatabaseCatalogResponse 终端用户数据库目录响应。
type DatabaseCatalogResponse struct {
	RequestID  string              `json:"requestId"`
	Databases  []*DatabaseLiteJSON `json:"databases"`
	TotalCount int64               `json:"totalCount"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	Error      *ErrorJSON          `json:"error,omitempty"`
}

// ModelCatalogResponse 终端用户模型目录响应（轻量字段）。
type ModelCatalogResponse struct {
	RequestID  string           `json:"requestId"`
	Models     []*ModelLiteJSON `json:"models"`
	TotalCount int64            `json:"totalCount"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	Error      *ErrorJSON       `json:"error,omitempty"`
}

// UpdateEndUserStatusRequest 更新状态请求
type UpdateEndUserStatusRequest struct {
	IsForbidden bool `json:"isForbidden"`
}

// UpdateEndUserStatusResponse 更新状态响应
type UpdateEndUserStatusResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
	Error     *ErrorJSON   `json:"error,omitempty"`
}

// DeleteEndUserResponse 删除响应
type DeleteEndUserResponse struct {
	RequestID string     `json:"requestId"`
	Success   bool       `json:"success"`
	Error     *ErrorJSON `json:"error,omitempty"`
}

// EndUserJSON 终端用户 JSON 表示
type EndUserJSON struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	IsForbidden bool    `json:"isForbidden"`
	CreatedBy   string  `json:"createdBy"`
	CreatedAt   *string `json:"createdAt,omitempty"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

// DatabaseLiteJSON 数据库简要信息（仅名称）。
type DatabaseLiteJSON struct {
	Name string `json:"name"`
}

// ModelLiteJSON 模型简要信息（不包含 fields/schema）。
type ModelLiteJSON struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	DatabaseName string `json:"databaseName"`
}

// PageInfoJSON 分页信息 JSON 表示
type PageInfoJSON struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

// AccessibleProjectJSON represents one accessible project for an end-user.
type AccessibleProjectJSON struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// GetAccessibleProjectsResponse 获取用户可访问项目列表的响应
type GetAccessibleProjectsResponse struct {
	RequestID string                   `json:"requestId"`
	Projects  []*AccessibleProjectJSON `json:"projects"`
	Error     *ErrorJSON               `json:"error,omitempty"`
}

// ErrorJSON 错误 JSON 表示
type ErrorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
