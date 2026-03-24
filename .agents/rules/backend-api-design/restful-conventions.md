---
paths:
  - "internal/interfaces/http/**/*.go"
  - "internal/app/**/*.go"
---

# API Design: RESTful Conventions

Enforce RESTful API design principles for HTTP endpoints including proper HTTP methods, status codes, request/response structure, and error handling.

**IMPORTANT**: RESTful APIs in this project are ONLY used for:
- User management endpoints
- Organization management endpoints
- Webhook/callback endpoints

All other features (projects, models, clusters, etc.) should use GraphQL APIs. See `graphql-patterns.md` for GraphQL design guidelines.

## Requirements

- **Use RESTful APIs ONLY for**: User management, Organization management, and Webhook/Callback endpoints
- **Use GraphQL for all other features**: Projects, Models, Clusters, Fields, Enums, etc.
- **Authentication**: User identity is passed via JWT token in `Authorization: Bearer <token>` header
- **Organization context**: Organization information is passed through URL path (e.g., `/api/orgs/{orgName}/...`)
- Use correct HTTP methods: GET (read), POST (create), PUT/PATCH (update), DELETE (remove)
- Return appropriate HTTP status codes: 200 (success), 201 (created), 204 (no content), 400 (bad request), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict), 500 (server error)
- Follow REST resource naming: use plural nouns (`/api/users`, `/api/orgs`, `/api/webhooks`) and avoid verbs in URLs
- **Use human-readable names (slugs) instead of numeric IDs in URLs** for multi-tenant resources (e.g., `/api/orgs/{name}` instead of `/api/orgs/{id}`)
- Organization and user identifiers must be URL-safe slugs: lowercase letters, numbers, hyphens, and underscores only
- Return consistent error responses using `bizerrors.BusinessError` with proper status code mapping
- Include `requestId` in all responses for traceability
- Use OpenAPI specification to document all endpoints in `api/openapi/*.yaml` modules

## Examples

### ✅ Good Example

```go
// RESTful endpoint for organization management (appropriate use case)
func (h *OrgHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var input models.CreateOrgInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        respondError(w, r, bizerrors.New(bizerrors.ParamInvalid, "invalid request body"))
        return
    }

    // Call application service
    org, err := h.service.CreateOrganization(r.Context(), &input)
    if err != nil {
        respondError(w, r, err)  // Proper error handling with status code mapping
        return
    }

    // Return 201 Created with resource
    respondJSON(w, http.StatusCreated, &models.CreateOrgResponse{
        RequestID: middleware.GetRequestID(r.Context()),
        Organization: org,
    })
}

// Get organization by name (slug) instead of numeric ID
func (h *OrgHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
    // Extract organization name from URL path
    orgName := chi.URLParam(r, "name")  // e.g., "acme-corp", "my-company"

    // User identity is already extracted from JWT token by auth middleware
    userID := middleware.GetUserID(r.Context())

    // Validate name format (lowercase, hyphens, underscores)
    if !isValidSlug(orgName) {
        respondError(w, r, bizerrors.New(bizerrors.ParamInvalid,
            "organization name must be lowercase with hyphens or underscores"))
        return
    }

    // Get organization by name (organization passed via path, user from token)
    org, err := h.service.GetOrganizationByName(r.Context(), orgName, userID)
    if err != nil {
        respondError(w, r, err)
        return
    }

    respondJSON(w, http.StatusOK, &models.GetOrgResponse{
        RequestID: middleware.GetRequestID(r.Context()),
        Organization: org,
    })
}

// Organization-scoped resource access
func (h *OrgHandler) GetOrganizationMembers(w http.ResponseWriter, r *http.Request) {
    // Organization context from path
    orgName := chi.URLParam(r, "name")

    // User identity from JWT token (extracted by middleware)
    userID := middleware.GetUserID(r.Context())

    // Get members (scoped to organization in path)
    members, err := h.service.GetOrgMembers(r.Context(), orgName, userID)
    if err != nil {
        respondError(w, r, err)
        return
    }

    respondJSON(w, http.StatusOK, &models.GetOrgMembersResponse{
        RequestID: middleware.GetRequestID(r.Context()),
        Members:   members,
    })
}

// Webhook callback endpoint (appropriate use case for REST)
func (h *WebhookHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    webhookID := chi.URLParam(r, "id")

    // Validate webhook signature
    if !h.validateSignature(r) {
        respondError(w, r, bizerrors.New(bizerrors.OperationDenied, "invalid signature"))
        return
    }

    // Process webhook payload
    if err := h.service.ProcessWebhook(r.Context(), webhookID, r.Body); err != nil {
        respondError(w, r, err)
        return
    }

    // Webhooks typically return 204 No Content on success
    w.WriteHeader(http.StatusNoContent)
}

// Helper: Validate slug format
func isValidSlug(s string) bool {
    // Lowercase letters, numbers, hyphens, underscores only
    match, _ := regexp.MatchString(`^[a-z0-9_-]+$`, s)
    return match && len(s) > 0 && len(s) <= 64
}

// Proper error response handling
func respondError(w http.ResponseWriter, r *http.Request, err error) {
    requestID := middleware.GetRequestID(r.Context())

    // Map business errors to HTTP status codes
    var bizErr *bizerrors.BusinessError
    if errors.As(err, &bizErr) {
        statusCode := mapErrorToStatusCode(bizErr)
        respondJSON(w, statusCode, &models.ErrorResponse{
            RequestID: requestID,
            Code:      bizErr.Info.GetCode(),
            Message:   bizErr.Msg(),
        })
        return
    }

    // Fallback to 500 for unexpected errors
    respondJSON(w, http.StatusInternalServerError, &models.ErrorResponse{
        RequestID: requestID,
        Code:      "SYSTEM_ERROR",
        Message:   "Internal server error",
    })
}

// Error to status code mapping
func mapErrorToStatusCode(err *bizerrors.BusinessError) int {
    switch err.Info.GetCode() {
    case bizerrors.NotFound.GetCode():
        return http.StatusNotFound
    case bizerrors.ParamInvalid.GetCode():
        return http.StatusBadRequest
    case bizerrors.Conflict.GetCode():
        return http.StatusConflict
    case bizerrors.OperationDenied.GetCode():
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}

// RESTful resource naming (good) - Use name/slug instead of ID
// User management (user identity from JWT token):
// GET    /api/users/me                    - Get current user (from token)
// PUT    /api/users/me                    - Update current user
// GET    /api/users/{id}                  - Get user by ID (admin only)

// Organization management (org in path, user from token):
// POST   /api/orgs                        - Create organization
// GET    /api/orgs/{orgName}              - Get org by name (e.g., /api/orgs/acme-corp)
// PUT    /api/orgs/{orgName}              - Update organization
// DELETE /api/orgs/{orgName}              - Delete organization
// GET    /api/orgs/{orgName}/members      - Get org members
// POST   /api/orgs/{orgName}/members      - Add org member

// Webhook/Callback endpoints:
// POST   /api/webhooks                    - Register webhook
// POST   /api/webhooks/{id}/callback      - Webhook callback endpoint
// DELETE /api/webhooks/{id}               - Delete webhook

// Example requests with authentication:
// GET /api/orgs/acme-corp
// Headers:
//   Authorization: Bearer eyJhbGc...  (JWT token contains user identity)
```

### ❌ Bad Example

```go
// ❌ Wrong: Using REST for project management (should use GraphQL)
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    // Projects, Models, Clusters should use GraphQL, not REST!
    // This is the wrong API design choice
}

// ❌ Wrong: Using REST for model CRUD (should use GraphQL)
func (h *ModelHandler) GetModels(w http.ResponseWriter, r *http.Request) {
    // Model management should be in GraphQL, not REST
}

// Poor REST API design (for appropriate REST resources)
func (h *OrgHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
    // Wrong: No input validation
    // Wrong: No request ID
    // Wrong: Generic error response
    // Wrong: Using 200 instead of 201 for creation

    org, err := h.service.CreateOrganization(r.Context(), nil)
    if err != nil {
        w.WriteHeader(http.StatusOK)  // Wrong status code!
        json.NewEncoder(w).Encode(map[string]string{
            "error": "something went wrong",  // No error code, no request ID
        })
        return
    }

    // Wrong: Should return 201, not 200
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(org)
}

// Wrong: Using numeric ID instead of name for multi-tenant resource
func (h *OrgHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
    // ❌ Bad: Using numeric ID in URL
    orgID := chi.URLParam(r, "id")  // e.g., "123"

    // ❌ Bad: Not validating slug format
    // ❌ Bad: Direct repository access instead of service
    org, err := h.repo.GetByID(r.Context(), orgID)
    if err != nil {
        w.WriteHeader(http.StatusOK)  // Wrong status code for error!
        return
    }

    json.NewEncoder(w).Encode(org)
}

// Bad resource naming (avoid verbs, avoid numeric IDs)
// POST /api/createOrg                  ❌ Has verb
// GET  /api/getOrg/{name}               ❌ Has verb
// POST /api/org/{name}/update           ❌ Wrong method + verb
// GET  /api/deleteOrg/{name}            ❌ Wrong method + verb
// GET  /api/orgs/123                    ❌ Using numeric ID instead of name (not multi-tenant friendly)
// GET  /api/orgs/My-Org                 ❌ Capital letters not allowed in slug
// GET  /api/orgs/my org                 ❌ Spaces not allowed in slug

// Wrong API choice - should use GraphQL instead
// POST /api/projects                    ❌ Projects should use GraphQL
// GET  /api/projects/{name}/models      ❌ Models should use GraphQL
// POST /api/clusters                    ❌ Clusters should use GraphQL
```

## Rationale

RESTful conventions improve API predictability, enable automatic client generation, and provide consistent error handling. Proper status codes and error responses help clients handle failures gracefully. OpenAPI documentation ensures API discoverability and maintainability.

**API Design Separation**: RESTful APIs are reserved for user management, organization management, and webhook/callback endpoints. All domain features (Projects, Models, Clusters, Fields, Enums) use GraphQL to provide flexible querying, type safety, and better client experience.

**Multi-tenant design with name-based routing**: Using human-readable names (slugs) instead of numeric IDs in URLs provides better multi-tenancy isolation, improves URL readability, and prevents resource enumeration attacks. Names are scoped to organizations, allowing different organizations to use the same resource names independently.

---

See skill: `backend-patterns` for comprehensive RESTful API design patterns and best practices.
