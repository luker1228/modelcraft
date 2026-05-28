package enduser

// CreateEndUserRequest is the request body for creating an end-user.
type CreateEndUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateEndUserStatusRequest is the request body for updating end-user status.
type UpdateEndUserStatusRequest struct {
	IsForbidden bool `json:"isForbidden"`
}

// EndUserJSON is the HTTP response shape for an end-user.
type EndUserJSON struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	IsForbidden bool    `json:"isForbidden"`
	CreatedAt   *string `json:"createdAt,omitempty"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

// PageInfoJSON contains cursor pagination info.
type PageInfoJSON struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

// CreateEndUserResponse is the response for a successful create request.
type CreateEndUserResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
}

// ListEndUsersResponse is the response for listing end-users.
type ListEndUsersResponse struct {
	RequestID  string         `json:"requestId"`
	Items      []*EndUserJSON `json:"items"`
	TotalCount int64          `json:"totalCount"`
	PageInfo   *PageInfoJSON  `json:"pageInfo,omitempty"`
}

// UpdateEndUserStatusResponse is the response for a successful status update.
type UpdateEndUserStatusResponse struct {
	RequestID string       `json:"requestId"`
	EndUser   *EndUserJSON `json:"endUser,omitempty"`
}

// DeleteEndUserResponse is the response for a successful delete request.
type DeleteEndUserResponse struct {
	RequestID string `json:"requestId"`
	Success   bool   `json:"success"`
}
