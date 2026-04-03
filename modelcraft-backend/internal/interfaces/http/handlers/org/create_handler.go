package org

import (
	"context"
	"encoding/json"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"regexp"
)

// CreateHandler handles organization creation requests.
type CreateHandler struct {
	createOrgService *organization.CreateOrganizationService
	logger           logfacade.Logger
}

// NewCreateHandler creates a new organization creation handler.
func NewCreateHandler(
	createOrgService *organization.CreateOrganizationService,
	logger logfacade.Logger,
) *CreateHandler {
	return &CreateHandler{
		createOrgService: createOrgService,
		logger:           logger,
	}
}

// Handle processes the organization initialization request.
func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req generated.InitOrganizationRequest

	// 1. Bind and validate request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf(r.Context(), "Invalid request body: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 2. Validate displayName length (max 255 characters)
	if len(req.DisplayName) > 255 {
		h.logger.Warnf(r.Context(), "Display name too long: %d characters", len(req.DisplayName))
		writeJSONError(w, http.StatusBadRequest, "Display name must not exceed 255 characters")
		return
	}

	// 3. Validate organization name format if provided (optional)
	orgName := ""
	if req.OrganizationName != nil {
		orgName = *req.OrganizationName
	}

	// 4. Extract user ID from context (set by JWT middleware)
	// This endpoint only accepts ModelCraft JWT, so user_id is the internal UUID
	userIDStr, err := ctxutils.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error(r.Context(), "User ID not found in context", logfacade.Err(err))
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	h.logger.Infof(
		r.Context(),
		"Creating organization: displayName=%s, orgName=%s, userID=%s",
		req.DisplayName, orgName, userIDStr,
	)

	// 5. Call Application Service
	output, err := h.createOrgService.Execute(r.Context(), &organization.CreateOrganizationInput{
		DisplayName:      req.DisplayName,
		OrganizationName: orgName,
		OwnerUserID:      userIDStr,
	})
	if err != nil {
		// 5. Error handling and HTTP response mapping
		statusCode, errMsg := h.mapErrorToHTTP(r.Context(), err)
		writeJSONError(w, statusCode, errMsg)
		return
	}

	// 6. Success response
	success := true
	alreadyExists := output.AlreadyExisted
	resp := generated.InitOrganizationResponse{
		Success:       &success,
		OrgName:       &output.OrganizationName,
		DisplayName:   &output.DisplayName,
		AlreadyExists: &alreadyExists,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// mapErrorToHTTP maps business errors to HTTP status codes
func (h *CreateHandler) mapErrorToHTTP(ctx context.Context, err error) (int, string) {
	if bizErr, ok := err.(*bizerrors.BusinessError); ok {
		switch bizErr.Info().GetCode() {
		case bizerrors.UserNotFound.GetCode():
			return http.StatusNotFound, "User not found. Please log in first."
		case bizerrors.OrganizationAlreadyExists.GetCode():
			return http.StatusConflict, "Organization name already exists. Please try a different name."
		case bizerrors.ParamInvalid.GetCode():
			return http.StatusBadRequest, bizErr.Msg()
		default:
			h.logger.Errorf(ctx, "Unhandled business error: %v", bizErr)
			return http.StatusInternalServerError, "Internal server error"
		}
	}
	h.logger.Errorf(ctx, "Unexpected error: %v", err)
	return http.StatusInternalServerError, "Internal server error"
}

// isValidOrgName validates the organization name format.
// Valid format: Slug with 6-24 characters
// - Must start with a lowercase letter [a-z]
// - Can contain lowercase letters, numbers, and underscores [a-z0-9_]
// - No hyphens allowed
// - Total length: 6-24 characters
//
// Examples:
// - Valid: "myorg", "my_company", "acme_corp_2024", "swiftlabs"
// - Invalid: "abc" (too short), "My-Company" (uppercase), "company-" (hyphen), "my-company" (hyphen)
func isValidOrgName(name string) bool {
	// Slug pattern: start with letter, only lowercase letters/numbers/underscores, 6-24 chars, no hyphens
	slugPattern := regexp.MustCompile(`^[a-z][a-z0-9_]{5,23}$`)
	return slugPattern.MatchString(name)
}
