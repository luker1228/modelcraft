package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/user"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Mock implementations ---

type mockUserRepo struct {
	users map[string]*user.User // keyed by external_id
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*user.User)}
}

func (m *mockUserRepo) Create(_ context.Context, u *user.User) error {
	if _, exists := m.users[u.ExternalID]; exists {
		return fmt.Errorf("duplicate external_id")
	}
	m.users[u.ExternalID] = u
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByExternalID(_ context.Context, externalID string) (*user.User, error) {
	return m.users[externalID], nil
}

func (m *mockUserRepo) ExistsByExternalID(_ context.Context, externalID string) (bool, error) {
	_, exists := m.users[externalID]
	return exists, nil
}

func (m *mockUserRepo) FindIDByExternalID(_ context.Context, externalID string) (string, bool, error) {
	if u, ok := m.users[externalID]; ok {
		return u.ID, true, nil
	}
	return "", false, nil
}

func (m *mockUserRepo) GetByPhone(_ context.Context, phone string) (*user.User, error) {
	for _, u := range m.users {
		if u.Phone.String() == phone {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) ExistsByPhone(_ context.Context, phone string) (bool, error) {
	for _, u := range m.users {
		if u.Phone.String() == phone {
			return true, nil
		}
	}
	return false, nil
}

// --- Helper functions ---

func makeWebhookRequest(handler http.HandlerFunc, payload interface{}, authHeader string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhook/casdoor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func parseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// --- Tests ---

func TestHandle_SignupWithExtendedUser(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "signup",
		User:         "luke1",
		Organization: "modelcraft",
		ExtendedUser: &CasdoorUserInfo{
			ID:          "casdoor-uuid-12345",
			Name:        "luke1",
			DisplayName: "Luke Skywalker",
			Email:       "luke1@example.com",
			Phone:       "13800138000",
			Owner:       "modelcraft",
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "")
	resp := parseResponse(w)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	if resp["message"] != "User created successfully" {
		t.Errorf("expected 'User created successfully', got %v", resp["message"])
	}

	if resp["userID"] == nil || resp["userID"] == "" {
		t.Error("expected userID in response")
	}

	// Verify user was created in repo with name and phone
	u, _ := repo.GetByExternalID(context.Background(), "casdoor-uuid-12345")
	if u == nil {
		t.Fatal("expected user to be created in repository")
	}
	if u.ExternalID != "casdoor-uuid-12345" {
		t.Errorf("expected external_id=casdoor-uuid-12345, got %s", u.ExternalID)
	}
	if u.Name != "Luke Skywalker" {
		t.Errorf("expected name='Luke Skywalker', got %s", u.Name)
	}
	if u.Phone.String() != "13800138000" {
		t.Errorf("expected phone='13800138000', got %s", u.Phone.String())
	}
}

func TestHandle_SignupIdempotent(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "signup",
		User:         "luke1",
		Organization: "modelcraft",
		ExtendedUser: &CasdoorUserInfo{
			ID:   "casdoor-uuid-12345",
			Name: "luke1",
		},
	}

	// First request: creates user
	w1 := makeWebhookRequest(handler.Handle, payload, "")
	if w1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", w1.Code)
	}
	resp1 := parseResponse(w1)
	firstUserID := resp1["userID"].(string)

	// Second request: should return existing user (idempotent)
	w2 := makeWebhookRequest(handler.Handle, payload, "")
	if w2.Code != http.StatusOK {
		t.Fatalf("second request: expected 200, got %d", w2.Code)
	}
	resp2 := parseResponse(w2)

	if resp2["message"] != "User already exists" {
		t.Errorf("expected 'User already exists', got %v", resp2["message"])
	}
	if resp2["userID"] != firstUserID {
		t.Errorf("expected same userID %s, got %v", firstUserID, resp2["userID"])
	}

	// Verify only one user in repo
	if len(repo.users) != 1 {
		t.Errorf("expected 1 user in repo, got %d", len(repo.users))
	}
}

func TestHandle_NonSignupEventIgnored(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "login",
		User:         "luke1",
		Organization: "modelcraft",
	}

	w := makeWebhookRequest(handler.Handle, payload, "")
	resp := parseResponse(w)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if resp["message"] != "Event ignored" {
		t.Errorf("expected 'Event ignored', got %v", resp["message"])
	}

	// Verify no user was created
	if len(repo.users) != 0 {
		t.Errorf("expected 0 users in repo, got %d", len(repo.users))
	}
}

func TestHandle_AddUserAction(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "add-user",
		User:         "admin1",
		Organization: "modelcraft",
		ExtendedUser: &CasdoorUserInfo{
			ID:   "admin-uuid-789",
			Name: "admin1",
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	resp := parseResponse(w)
	if resp["message"] != "User created successfully" {
		t.Errorf("expected 'User created successfully', got %v", resp["message"])
	}
}

func TestHandle_WebhookSecretValid(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "my-secret-123")

	payload := CasdoorWebhookPayload{
		Action: "signup",
		User:   "luke1",
		ExtendedUser: &CasdoorUserInfo{
			ID:   "casdoor-uuid-999",
			Name: "luke1",
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "my-secret-123")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandle_WebhookSecretInvalid(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "my-secret-123")

	payload := CasdoorWebhookPayload{
		Action: "signup",
		User:   "luke1",
		ExtendedUser: &CasdoorUserInfo{
			ID:   "casdoor-uuid-999",
			Name: "luke1",
		},
	}

	// Wrong secret
	w := makeWebhookRequest(handler.Handle, payload, "wrong-secret")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	// Missing secret
	w2 := makeWebhookRequest(handler.Handle, payload, "")

	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for missing secret, got %d", w2.Code)
	}

	// Verify no user was created
	if len(repo.users) != 0 {
		t.Errorf("expected 0 users in repo, got %d", len(repo.users))
	}
}

func TestHandle_WebhookSecretNotConfigured(t *testing.T) {
	repo := newMockUserRepo()
	// No webhook secret configured - should accept all requests
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action: "signup",
		User:   "luke1",
		ExtendedUser: &CasdoorUserInfo{
			ID:   "casdoor-uuid-abc",
			Name: "luke1",
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 when no secret configured, got %d", w.Code)
	}
}

func TestHandle_InvalidPayload(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/webhook/casdoor", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Handle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandle_SignupWithoutExtendedUser(t *testing.T) {
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	// Without extendedUser, we cannot determine the Casdoor UUID
	payload := CasdoorWebhookPayload{
		Action:       "signup",
		User:         "luke1",
		Organization: "modelcraft",
		ExtendedUser: nil,
	}

	w := makeWebhookRequest(handler.Handle, payload, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when extendedUser is missing, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseResponse(w)
	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		t.Error("expected error message about missing extendedUser")
	}
}

func TestHandle_RealCasdoorPayloadWithoutExtendedUser(t *testing.T) {
	// Real Casdoor payload where "Is User Extended" is NOT enabled
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	objectJSON := `{"application":"modelcraft","organization":"modelcraft","username":"luke1",` +
		`"name":"luke1","password":"***","confirm":"123456","countryCode":"US","agreement":true}`
	rawPayload := strings.Join([]string{
		"{",
		"  \"id\": 0,",
		"  \"owner\": \"modelcraft\",",
		"  \"name\": \"dc64e4d0-76a1-4814-ab4a-e448bea71505\",",
		"  \"createdTime\": \"2026-02-10T13:38:27Z\",",
		"  \"organization\": \"modelcraft\",",
		"  \"clientIp\": \"11.176.18.8\",",
		"  \"user\": \"luke1\",",
		"  \"method\": \"POST\",",
		"  \"requestUri\": \"/api/signup\",",
		"  \"action\": \"signup\",",
		"  \"language\": \"zh\",",
		fmt.Sprintf("  \"object\": \"%s\",", objectJSON),
		"  \"response\": \"{status:\\\"ok\\\", msg:\\\"\\\"}\",",
		"  \"statusCode\": 200,",
		"  \"isTriggered\": false,",
		"  \"extendedUser\": null",
		"}",
	}, "\n")

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/casdoor", bytes.NewReader([]byte(rawPayload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Handle(w, req)

	// Without extendedUser, we cannot resolve the Casdoor UUID
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 (no extendedUser), got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandle_RealCasdoorPayloadWithExtendedUser(t *testing.T) {
	// Test the case where Casdoor has "Is User Extended" enabled
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	rawPayload := `{
		"id": 0,
		"owner": "modelcraft",
		"name": "dc64e4d0-76a1-4814-ab4a-e448bea71505",
		"organization": "modelcraft",
		"user": "luke1",
		"action": "signup",
		"object": "{}",
		"extendedUser": {
			"id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			"name": "luke1",
			"email": "luke1@example.com",
			"owner": "modelcraft"
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/webhook/casdoor", bytes.NewReader([]byte(rawPayload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify user was created with the correct external ID
	u, _ := repo.GetByExternalID(context.Background(), "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if u == nil {
		t.Fatal("expected user to be created with extendedUser.id")
	}
}

func TestHandle_ProfileFromObjectJSON(t *testing.T) {
	// extendedUser provides ID but no displayName/phone; object JSON fills the gaps
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "signup",
		User:         "luke1",
		Organization: "modelcraft",
		Object:       `{"username":"luke1","name":"Luke S","phone":"13900139000"}`,
		ExtendedUser: &CasdoorUserInfo{
			ID: "casdoor-uuid-obj-test",
			// No DisplayName or Phone in extendedUser
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	u, _ := repo.GetByExternalID(context.Background(), "casdoor-uuid-obj-test")
	if u == nil {
		t.Fatal("expected user to be created")
	}
	if u.Name != "Luke S" {
		t.Errorf("expected name='Luke S' from object JSON, got %s", u.Name)
	}
	if u.Phone.String() != "13900139000" {
		t.Errorf("expected phone='13900139000' from object JSON, got %s", u.Phone.String())
	}
}

func TestHandle_NameFallbackToUsername(t *testing.T) {
	// extendedUser has ID only, object has no name — should fall back to username
	repo := newMockUserRepo()
	handler := NewCasdoorHandler(repo, "")

	payload := CasdoorWebhookPayload{
		Action:       "signup",
		User:         "luke1",
		Organization: "modelcraft",
		ExtendedUser: &CasdoorUserInfo{
			ID: "casdoor-uuid-fallback",
		},
	}

	w := makeWebhookRequest(handler.Handle, payload, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	u, _ := repo.GetByExternalID(context.Background(), "casdoor-uuid-fallback")
	if u == nil {
		t.Fatal("expected user to be created")
	}
	if u.Name != "luke1" {
		t.Errorf("expected name='luke1' (username fallback), got %s", u.Name)
	}
}

func TestIsUserSignupAction(t *testing.T) {
	tests := []struct {
		action   string
		expected bool
	}{
		{"signup", true},
		{"add-user", true},
		{"login", false},
		{"logout", false},
		{"update-user", false},
		{"delete-user", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := isUserSignupAction(tt.action)
			if result != tt.expected {
				t.Errorf("isUserSignupAction(%q) = %v, want %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestResolveUsername(t *testing.T) {
	// Test with User field
	payload1 := &CasdoorWebhookPayload{User: "luke1"}
	if got := resolveUsername(payload1); got != "luke1" {
		t.Errorf("expected 'luke1', got %q", got)
	}

	// Test with Object field fallback
	payload2 := &CasdoorWebhookPayload{
		User:   "",
		Object: `{"username":"john","organization":"myorg"}`,
	}
	if got := resolveUsername(payload2); got != "john" {
		t.Errorf("expected 'john', got %q", got)
	}

	// Test with empty fields
	payload3 := &CasdoorWebhookPayload{User: "", Object: ""}
	if got := resolveUsername(payload3); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}

	// Test with invalid Object JSON
	payload4 := &CasdoorWebhookPayload{User: "", Object: "invalid json"}
	if got := resolveUsername(payload4); got != "" {
		t.Errorf("expected empty string for invalid JSON, got %q", got)
	}
}

func TestResolveExternalID(t *testing.T) {
	// With extendedUser
	payload1 := &CasdoorWebhookPayload{
		ExtendedUser: &CasdoorUserInfo{ID: "uuid-123"},
	}
	if got := resolveExternalID(payload1); got != "uuid-123" {
		t.Errorf("expected 'uuid-123', got %q", got)
	}

	// Without extendedUser
	payload2 := &CasdoorWebhookPayload{ExtendedUser: nil}
	if got := resolveExternalID(payload2); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}

	// With extendedUser but empty ID
	payload3 := &CasdoorWebhookPayload{
		ExtendedUser: &CasdoorUserInfo{ID: ""},
	}
	if got := resolveExternalID(payload3); got != "" {
		t.Errorf("expected empty string for empty ID, got %q", got)
	}
}

func TestResolveUserProfile(t *testing.T) {
	// Full extendedUser
	payload1 := &CasdoorWebhookPayload{
		ExtendedUser: &CasdoorUserInfo{
			DisplayName: "Luke Skywalker",
			Name:        "luke1",
			Phone:       "13800138000",
		},
	}
	name, phone := resolveUserProfile(payload1)
	if name != "Luke Skywalker" {
		t.Errorf("expected 'Luke Skywalker', got %q", name)
	}
	if phone != "13800138000" {
		t.Errorf("expected '13800138000', got %q", phone)
	}

	// extendedUser without displayName — falls back to Name
	payload2 := &CasdoorWebhookPayload{
		ExtendedUser: &CasdoorUserInfo{
			Name:  "luke1",
			Phone: "13800138000",
		},
	}
	name2, _ := resolveUserProfile(payload2)
	if name2 != "luke1" {
		t.Errorf("expected 'luke1', got %q", name2)
	}

	// No extendedUser, object JSON provides info
	payload3 := &CasdoorWebhookPayload{
		User:   "luke1",
		Object: `{"name":"Luke S","phone":"13900139000"}`,
	}
	name3, phone3 := resolveUserProfile(payload3)
	if name3 != "Luke S" {
		t.Errorf("expected 'Luke S', got %q", name3)
	}
	if phone3 != "13900139000" {
		t.Errorf("expected '13900139000', got %q", phone3)
	}

	// No extendedUser, no object — falls back to username
	payload4 := &CasdoorWebhookPayload{
		User: "luke1",
	}
	name4, phone4 := resolveUserProfile(payload4)
	if name4 != "luke1" {
		t.Errorf("expected 'luke1' (username fallback), got %q", name4)
	}
	if phone4 != "" {
		t.Errorf("expected empty phone, got %q", phone4)
	}
}
