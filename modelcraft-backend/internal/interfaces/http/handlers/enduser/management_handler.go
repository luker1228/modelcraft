package enduser

import (
	"encoding/json"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	appEnduser "modelcraft/internal/app/enduser"
)

// ManagementHandler 处理终端用户管理 HTTP 请求
type ManagementHandler struct {
	appService *appEnduser.EndUserManagementAppService
	logger     logfacade.Logger
}

const errMsgOrgProjectRequired = "X-Org-Name and X-Project-Slug headers are required"

// NewManagementHandler 创建管理 Handler
func NewManagementHandler(
	appService *appEnduser.EndUserManagementAppService,
	logger logfacade.Logger,
) *ManagementHandler {
	return &ManagementHandler{
		appService: appService,
		logger:     logger,
	}
}

// Create 处理 POST /internal/end-users
func (h *ManagementHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	// 1. 解析请求体
	var req CreateEndUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(ctx, "Invalid create end user request body", logfacade.Err(err))
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	// 2. 从 Header 提取 orgName 和 projectSlug
	orgName := r.Header.Get(httpheader.XOrgName)
	projectSlug := r.Header.Get(httpheader.XProjectSlug)
	createdBy := r.Header.Get(httpheader.XUserID) // 开发者 ID

	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", errMsgOrgProjectRequired)
		return
	}

	// 3. 调用 App 层
	result, err := h.appService.CreateEndUser(ctx, appEnduser.CreateEndUserCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Username:    req.Username,
		Password:    req.Password,
		CreatedBy:   createdBy,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Create end user failed")
		return
	}

	// 4. 返回成功响应
	createdAt := result.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	updatedAt := result.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	h.writeJSON(w, http.StatusCreated, CreateEndUserResponse{
		RequestID: requestID,
		EndUser: &EndUserJSON{
			ID:          result.ID,
			Username:    result.Username,
			IsForbidden: result.IsForbidden,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		},
	})
}

// List 处理 GET /internal/end-users
func (h *ManagementHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	// 1. 从 Header 提取 orgName 和 projectSlug
	orgName := r.Header.Get(httpheader.XOrgName)
	projectSlug := r.Header.Get(httpheader.XProjectSlug)

	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", errMsgOrgProjectRequired)
		return
	}

	// 2. 解析查询参数
	query := r.URL.Query()
	search := query.Get("search")
	after := query.Get("after")
	first := 20
	if firstStr := query.Get("first"); firstStr != "" {
		if f, err := strconv.Atoi(firstStr); err == nil && f > 0 {
			first = f
		}
	}

	// 3. 调用 App 层
	result, err := h.appService.ListEndUsers(ctx, appEnduser.ListEndUsersCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Search:      search,
		First:       first,
		After:       after,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "List end users failed")
		return
	}

	// 4. 转换结果
	items := make([]*EndUserJSON, 0, len(result.Items))
	for _, u := range result.Items {
		createdAt := u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
		updatedAt := u.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
		items = append(items, &EndUserJSON{
			ID:          u.ID,
			Username:    u.Username,
			IsForbidden: u.IsForbidden,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		})
	}

	var endCursor *string
	if result.EndCursor != "" {
		endCursor = &result.EndCursor
	}

	h.writeJSON(w, http.StatusOK, ListEndUsersResponse{
		RequestID:  requestID,
		Items:      items,
		TotalCount: result.TotalCount,
		PageInfo: &PageInfoJSON{
			HasNextPage: result.HasNextPage,
			EndCursor:   endCursor,
		},
	})
}

// UpdateStatus 处理 PATCH /internal/end-users/{userId}/status
func (h *ManagementHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	// 1. 从 URL 提取 userId
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "userId is required")
		return
	}

	// 2. 解析请求体
	var req UpdateEndUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(ctx, "Invalid update status request body", logfacade.Err(err))
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "Invalid request body")
		return
	}

	// 3. 从 Header 提取 orgName 和 projectSlug
	orgName := r.Header.Get(httpheader.XOrgName)
	projectSlug := r.Header.Get(httpheader.XProjectSlug)

	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", errMsgOrgProjectRequired)
		return
	}

	// 4. 调用 App 层
	result, err := h.appService.UpdateEndUserStatus(ctx, appEnduser.UpdateEndUserStatusCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		UserID:      userID,
		IsForbidden: req.IsForbidden,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Update end user status failed")
		return
	}

	// 5. 返回成功响应
	createdAt := result.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	updatedAt := result.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	h.writeJSON(w, http.StatusOK, UpdateEndUserStatusResponse{
		RequestID: requestID,
		EndUser: &EndUserJSON{
			ID:          result.ID,
			Username:    result.Username,
			IsForbidden: result.IsForbidden,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		},
	})
}

// Delete 处理 DELETE /internal/end-users/{userId}
func (h *ManagementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	// 1. 从 URL 提取 userId
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "userId is required")
		return
	}

	// 2. 从 Header 提取 orgName 和 projectSlug
	orgName := r.Header.Get(httpheader.XOrgName)
	projectSlug := r.Header.Get(httpheader.XProjectSlug)

	if orgName == "" || projectSlug == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", errMsgOrgProjectRequired)
		return
	}

	// 3. 调用 App 层
	err := h.appService.DeleteEndUser(ctx, appEnduser.DeleteEndUserCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		UserID:      userID,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "Delete end user failed")
		return
	}

	// 4. 返回成功响应
	h.writeJSON(w, http.StatusOK, DeleteEndUserResponse{
		RequestID: requestID,
		Success:   true,
	})
}

// ---------------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------------

func (h *ManagementHandler) handleBusinessError(
	w http.ResponseWriter,
	r *http.Request,
	requestID string,
	err error,
	logMsg string,
) {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		// Not a BusinessError — unexpected system error
		h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
		h.writeError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	// Log with stack at the Interfaces layer error conversion point
	h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))

	statusCode := bizErr.GetHTTPStatusCode()
	code := bizErr.Info().GetCode()
	msg := bizErr.Msg()

	h.writeError(w, statusCode, requestID, code, msg)
}

func (h *ManagementHandler) writeError(w http.ResponseWriter, statusCode int, requestID, code, message string) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requestId": requestID,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func (h *ManagementHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set(httpheader.ContentType, httpheader.ContentTypeApplicationJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
