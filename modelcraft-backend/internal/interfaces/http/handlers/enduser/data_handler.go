package enduser

import (
	"encoding/json"
	"net/http"
	"strconv"

	appModelDesign "modelcraft/internal/app/modeldesign"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
)

// DataHandler handles end-user data metadata APIs.
type DataHandler struct {
	modelService *appModelDesign.ModelDesignAppService
	logger       logfacade.Logger
}

func NewDataHandler(
	modelService *appModelDesign.ModelDesignAppService,
	logger logfacade.Logger,
) *DataHandler {
	return &DataHandler{
		modelService: modelService,
		logger:       logger,
	}
}

// DatabaseCatalog handles GET /internal/end-user/data/database-catalog.
func (h *DataHandler) DatabaseCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	endUserID := r.Header.Get("X-End-User-Id")
	if orgName == "" || projectSlug == "" || endUserID == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "X-Org-Name, X-Project-Slug and X-End-User-Id headers are required")
		return
	}

	query := r.URL.Query()
	search := query.Get("search")
	page := 1
	pageSize := 20
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "page must be a positive integer")
			return
		}
		page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "pageSize must be a positive integer")
			return
		}
		pageSize = parsed
	}

	databases, totalCount, err := h.modelService.QueryDatabaseCatalogWithCommand(ctx, appModelDesign.DatabaseCatalogQueryCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Search:      search,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user database catalog failed")
		return
	}

	items := make([]*DatabaseLiteJSON, 0, len(databases))
	for _, db := range databases {
		items = append(items, &DatabaseLiteJSON{Name: db})
	}

	h.writeJSON(w, http.StatusOK, DatabaseCatalogResponse{
		RequestID:  requestID,
		Databases:  items,
		TotalCount: int64(totalCount),
		Page:       page,
		PageSize:   pageSize,
	})
}

// ModelCatalog handles GET /internal/end-user/data/model-catalog.
func (h *DataHandler) ModelCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := ctxutils.GetRequestID(ctx)

	orgName := r.Header.Get("X-Org-Name")
	projectSlug := r.Header.Get("X-Project-Slug")
	endUserID := r.Header.Get("X-End-User-Id")
	if orgName == "" || projectSlug == "" || endUserID == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "X-Org-Name, X-Project-Slug and X-End-User-Id headers are required")
		return
	}

	query := r.URL.Query()
	databaseName := query.Get("databaseName")
	if databaseName == "" {
		h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "databaseName is required")
		return
	}
	search := query.Get("search")
	page := 1
	pageSize := 50
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "page must be a positive integer")
			return
		}
		page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.writeError(w, http.StatusBadRequest, requestID, "PARAM_INVALID", "pageSize must be a positive integer")
			return
		}
		pageSize = parsed
	}

	models, totalCount, err := h.modelService.QueryModelsWithCommand(ctx, appModelDesign.ModelQueryCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
		Name:         search,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		h.handleBusinessError(w, r, requestID, err, "End-user model catalog failed")
		return
	}

	items := make([]*ModelLiteJSON, 0, len(models))
	for _, model := range models {
		items = append(items, &ModelLiteJSON{
			ID:           model.ID,
			Name:         model.ModelName,
			Title:        model.Title,
			DatabaseName: model.DatabaseName,
		})
	}

	h.writeJSON(w, http.StatusOK, ModelCatalogResponse{
		RequestID:  requestID,
		Models:     items,
		TotalCount: int64(totalCount),
		Page:       page,
		PageSize:   pageSize,
	})
}

func (h *DataHandler) handleBusinessError(
	w http.ResponseWriter,
	r *http.Request,
	requestID string,
	err error,
	logMsg string,
) {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
		h.writeError(w, http.StatusInternalServerError, requestID, "SYSTEM_ERROR", "Internal server error")
		return
	}

	h.logger.Error(r.Context(), logMsg, logfacade.Err(err), logfacade.Stack(err))
	h.writeError(w, bizErr.GetHTTPStatusCode(), requestID, bizErr.Info().GetCode(), bizErr.Msg())
}

func (h *DataHandler) writeError(w http.ResponseWriter, statusCode int, requestID, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requestId": requestID,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func (h *DataHandler) writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
