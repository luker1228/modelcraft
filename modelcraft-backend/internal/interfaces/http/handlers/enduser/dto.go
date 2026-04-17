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
	RequestID string        `json:"requestId"`
	EndUser   *EndUserJSON  `json:"endUser,omitempty"`
	Error     *ErrorJSON    `json:"error,omitempty"`
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

// PageInfoJSON 分页信息 JSON 表示
type PageInfoJSON struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

// ErrorJSON 错误 JSON 表示
type ErrorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
