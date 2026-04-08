package webhook

import (
	"encoding/json"
	"modelcraft/internal/domain/user"
	"modelcraft/pkg/logfacade"
	"net/http"

	"github.com/google/uuid"
)

// CasdoorWebhookPayload represents the actual webhook payload sent by Casdoor.
// Casdoor sends the Record fields at the top level (flat structure), not nested.
//
// Example payload from Casdoor:
//
//	{
//	  "id": 0,
//	  "owner": "modelcraft",
//	  "name": "dc64e4d0-...",
//	  "organization": "modelcraft",
//	  "user": "luke1",
//	  "action": "signup",
//	  "object": "{\"username\":\"luke1\",...}",
//	  "extendedUser": { "id": "...", "name": "luke1", ... }
//	}
type CasdoorWebhookPayload struct {
	ID           int    `json:"id"`
	Owner        string `json:"owner"`
	Name         string `json:"name"` // Record UUID (NOT user ID)
	Organization string `json:"organization"`
	User         string `json:"user"`   // Username (e.g., "luke1")
	Action       string `json:"action"` // Event type: "signup", "login", etc.
	Object       string `json:"object"` // JSON string of the request body
	Response     string `json:"response"`
	// Full user info (requires "Is User Extended" in Casdoor webhook config)
	ExtendedUser *CasdoorUserInfo `json:"extendedUser"`
}

// CasdoorUserInfo represents the extended user object in the webhook payload.
// Present when "Is User Extended" is enabled in the Casdoor webhook config.
// This is the primary source of user identity information.
type CasdoorUserInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Owner       string `json:"owner"`
}

// SignupObject represents the parsed "object" field for signup events.
// Used as a fallback source for user profile information (name, phone).
type SignupObject struct {
	Application  string `json:"application"`
	Organization string `json:"organization"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
}

// CasdoorHandler handles Casdoor webhook events.
// It relies solely on the webhook payload data — no Casdoor API calls are made.
// Requires "Is User Extended" to be enabled in the Casdoor webhook configuration
// so that extendedUser is populated with the user's Casdoor UUID.
type CasdoorHandler struct {
	userRepo      user.UserRepository
	webhookSecret string
}

// NewCasdoorHandler creates a new Casdoor webhook handler.
func NewCasdoorHandler(
	userRepo user.UserRepository,
	webhookSecret string,
) *CasdoorHandler {
	return &CasdoorHandler{
		userRepo:      userRepo,
		webhookSecret: webhookSecret,
	}
}

// Handle processes incoming Casdoor webhook events.
// POST /api/webhook/casdoor
//
// When a user registers in Casdoor, this endpoint creates the corresponding
// user record in ModelCraft's users table. The flow:
//  1. Parse the flat webhook payload from Casdoor
//  2. Filter for "signup" / "add-user" actions only
//  3. Extract the user's Casdoor ID from extendedUser (requires "Is User Extended" enabled)
//  4. Extract profile info (name, phone) from extendedUser or object JSON
//  5. Create a record in the users table (idempotent)
func (h *CasdoorHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logfacade.GetLogger(ctx)
	// Verify webhook secret if configured
	if h.webhookSecret != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != h.webhookSecret {
			logger.Warnf(ctx, "Webhook request rejected: invalid authorization header")
			writeJSONError(w, http.StatusUnauthorized, "Invalid webhook secret")
			return
		}
	}

	// Parse webhook payload
	var payload CasdoorWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Warnf(ctx, "Failed to parse Casdoor webhook payload: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	logger.Infof(ctx, "Received Casdoor webhook: action=%s, user=%s, org=%s",
		payload.Action, payload.User, payload.Organization)

	// Only process user signup events
	if !isUserSignupAction(payload.Action) {
		logger.Infof(ctx, "Ignoring non-signup webhook event: action=%s", payload.Action)
		writeJSONResponse(w, http.StatusOK, map[string]string{
			"message": "Event ignored",
			"action":  payload.Action,
		})
		return
	}

	// Extract external ID from extendedUser — this is required.
	// If extendedUser is nil, "Is User Extended" is not enabled in Casdoor webhook config.
	externalID := resolveExternalID(&payload)
	if externalID == "" {
		logger.Warnf(ctx,
			"Missing extendedUser in webhook payload; enable 'Is User Extended' in Casdoor webhook config")
		writeJSONError(w, http.StatusBadRequest,
			"Cannot determine user ID: enable 'Is User Extended' in Casdoor webhook configuration")
		return
	}

	// Check if user already exists (idempotent)
	existingUserID, found, err := h.userRepo.FindIDByExternalID(r.Context(), externalID)
	if err != nil {
		logger.Error(ctx, "Failed to check existing user", logfacade.Err(err))
		writeJSONError(w, http.StatusInternalServerError, "Failed to check existing user")
		return
	}

	if found {
		logger.Infof(ctx, "User already exists in ModelCraft DB: external_id=%s, skipping creation", externalID)
		writeJSONResponse(w, http.StatusOK, map[string]string{
			"message": "User already exists",
			"userID":  existingUserID,
		})
		return
	}

	// Resolve name and phone from webhook payload
	name, phone := resolveUserProfile(&payload)

	// Create user record
	userID := uuid.New().String()
	newUser, err := user.NewOAuthUser(userID, externalID, name, phone)
	if err != nil {
		logger.Error(ctx, "Failed to create user entity", logfacade.Err(err))
		writeJSONError(w, http.StatusInternalServerError, "Failed to create user entity")
		return
	}

	if err := h.userRepo.Create(r.Context(), newUser); err != nil {
		logger.Error(ctx, "Failed to save user to database", logfacade.Err(err))
		writeJSONError(w, http.StatusInternalServerError, "Failed to save user")
		return
	}

	logger.Infof(ctx, "User created via webhook: id=%s, external_id=%s, username=%s", userID, externalID, payload.User)
	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User created successfully",
		"userID":  userID,
	})
}

// resolveExternalID extracts the user's Casdoor UUID from the webhook payload.
// The only reliable source is extendedUser.id, which requires "Is User Extended"
// to be enabled in the Casdoor webhook configuration.
func resolveExternalID(payload *CasdoorWebhookPayload) string {
	if payload.ExtendedUser != nil && payload.ExtendedUser.ID != "" {
		return payload.ExtendedUser.ID
	}
	return ""
}

// resolveUsername extracts the username from the webhook payload.
// It first tries payload.User, then falls back to parsing the object JSON.
func resolveUsername(payload *CasdoorWebhookPayload) string {
	if payload.User != "" {
		return payload.User
	}

	if payload.Object != "" {
		var obj SignupObject
		if err := json.Unmarshal([]byte(payload.Object), &obj); err == nil && obj.Username != "" {
			return obj.Username
		}
	}

	return ""
}

// resolveUserProfile extracts name and phone from the webhook payload.
// Priority: extendedUser fields > object JSON fields > username as fallback name.
func resolveUserProfile(payload *CasdoorWebhookPayload) (name, phone string) {
	name, phone = resolveExtendedUserProfile(payload)
	name, phone = fillProfileFromSignupObject(name, phone, payload.Object)

	// Strategy 3: Fall back to username as name
	if name == "" {
		name = resolveUsername(payload)
	}

	return name, phone
}

func resolveExtendedUserProfile(payload *CasdoorWebhookPayload) (string, string) {
	if payload.ExtendedUser == nil {
		return "", ""
	}

	name := payload.ExtendedUser.DisplayName
	if name == "" {
		name = payload.ExtendedUser.Name
	}
	phone := payload.ExtendedUser.Phone
	return name, phone
}

func fillProfileFromSignupObject(name, phone, object string) (string, string) {
	if name != "" && phone != "" {
		return name, phone
	}

	obj, ok := parseSignupObject(object)
	if !ok {
		return name, phone
	}

	if name == "" && obj.Name != "" {
		name = obj.Name
	}
	if phone == "" && obj.Phone != "" {
		phone = obj.Phone
	}
	return name, phone
}

func parseSignupObject(object string) (*SignupObject, bool) {
	if object == "" {
		return nil, false
	}

	var obj SignupObject
	if err := json.Unmarshal([]byte(object), &obj); err != nil {
		return nil, false
	}
	return &obj, true
}

// isUserSignupAction checks whether the webhook action represents a user signup event.
func isUserSignupAction(action string) bool {
	switch action {
	case "signup", "add-user":
		return true
	default:
		return false
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
