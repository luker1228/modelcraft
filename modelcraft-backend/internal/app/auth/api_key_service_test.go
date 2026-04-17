package auth

import (
	"context"
	"errors"
	"modelcraft/pkg/bizerrors"
	"strings"
	"testing"
	"time"

	domainauth "modelcraft/internal/domain/auth"
)

// mockAPIKeyRepo is an in-memory mock of domainauth.APIKeyRepository.
type mockAPIKeyRepo struct {
	keys     map[string]*domainauth.APIKey
	byHash   map[string]*domainauth.APIKey
	countErr error
	saveErr  error
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{
		keys:   make(map[string]*domainauth.APIKey),
		byHash: make(map[string]*domainauth.APIKey),
	}
}

func (m *mockAPIKeyRepo) Save(ctx context.Context, key *domainauth.APIKey) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.keys[key.ID] = key
	m.byHash[key.KeyHash] = key
	return nil
}

func (m *mockAPIKeyRepo) FindByHash(ctx context.Context, hash string) (*domainauth.APIKey, error) {
	k := m.byHash[hash]
	return k, nil
}

func (m *mockAPIKeyRepo) FindByID(ctx context.Context, id string) (*domainauth.APIKey, error) {
	k := m.keys[id]
	return k, nil
}

func (m *mockAPIKeyRepo) ListByUserID(ctx context.Context, userID string) ([]*domainauth.APIKey, error) {
	var result []*domainauth.APIKey
	for _, k := range m.keys {
		if k.UserID == userID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepo) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	count := 0
	for _, k := range m.keys {
		if k.UserID == userID && k.RevokedAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *mockAPIKeyRepo) Revoke(ctx context.Context, id, userID string) error {
	k := m.keys[id]
	if k == nil {
		return nil
	}
	now := time.Now()
	k.RevokedAt = &now
	return nil
}

func (m *mockAPIKeyRepo) Update(ctx context.Context, id, userID, name string, roleIDs []int, expiresAt *time.Time) error {
	k := m.keys[id]
	if k == nil {
		return nil
	}
	k.Name = name
	k.RoleIDs = roleIDs
	k.ExpiresAt = expiresAt
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	return nil
}

func (m *mockAPIKeyRepo) DeleteRevoked(ctx context.Context) error {
	return nil
}

// Compile-time check
var _ domainauth.APIKeyRepository = (*mockAPIKeyRepo)(nil)

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestAPIKeyService_CreateAPIKey_Success(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)

	result, err := svc.CreateAPIKey(context.Background(), CreateAPIKeyCommand{
		UserID: "user1",
		Name:   "my-key",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !strings.HasPrefix(result.PlainKey, "mc_") {
		t.Errorf("expected key to start with mc_, got: %s", result.PlainKey)
	}
	if result.Key.UserID != "user1" {
		t.Errorf("expected userID user1, got: %s", result.Key.UserID)
	}
	if result.Key.Name != "my-key" {
		t.Errorf("expected name my-key, got: %s", result.Key.Name)
	}
	if result.Key.KeyPrefix != result.PlainKey[:10] {
		t.Errorf("key prefix mismatch: %s vs %s", result.Key.KeyPrefix, result.PlainKey[:10])
	}
}

func TestAPIKeyService_CreateAPIKey_LimitExceeded(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	// Pre-fill 20 keys
	for i := 0; i < domainauth.APIKeyMaxPerUser; i++ {
		_, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{
			UserID: "user1",
			Name:   "key",
		})
		if err != nil {
			t.Fatalf("setup: unexpected error on key %d: %v", i, err)
		}
	}

	// 21st should fail
	_, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{
		UserID: "user1",
		Name:   "overflow-key",
	})
	if err == nil {
		t.Fatal("expected error for limit exceeded, got nil")
	}
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		t.Fatalf("expected *bizerrors.BusinessError, got %T", err)
	}
	if bizErr.Info().GetCode() != bizerrors.APIKeyLimitExceeded.GetCode() {
		t.Errorf("expected APIKeyLimitExceeded code, got: %s", bizErr.Info().GetCode())
	}
}

func TestAPIKeyService_RevokeAPIKey_Success(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	// Create a key first
	result, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "k"})
	if err != nil {
		t.Fatal(err)
	}

	revoked, err := svc.RevokeAPIKey(ctx, result.Key.ID, "user1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if revoked == nil {
		t.Fatal("expected revoked key, got nil")
	}
	if revoked.RevokedAt == nil {
		t.Error("expected RevokedAt to be set")
	}
}

func TestAPIKeyService_RevokeAPIKey_NotFound(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)

	_, err := svc.RevokeAPIKey(context.Background(), "nonexistent-id", "user1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		t.Fatalf("expected *bizerrors.BusinessError, got %T", err)
	}
	if bizErr.Info().GetCode() != bizerrors.APIKeyNotFound.GetCode() {
		t.Errorf("expected APIKeyNotFound code, got: %s", bizErr.Info().GetCode())
	}
}

func TestAPIKeyService_RevokeAPIKey_WrongUser(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	result, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "k"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.RevokeAPIKey(ctx, result.Key.ID, "user2")
	if err == nil {
		t.Fatal("expected error for wrong user, got nil")
	}
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		t.Fatalf("expected *bizerrors.BusinessError, got %T", err)
	}
	if bizErr.Info().GetCode() != bizerrors.APIKeyNotFound.GetCode() {
		t.Errorf("expected APIKeyNotFound code, got: %s", bizErr.Info().GetCode())
	}
}

func TestAPIKeyService_UpdateAPIKey_Success(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	result, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "old-name"})
	if err != nil {
		t.Fatal(err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	updated, err := svc.UpdateAPIKey(ctx, UpdateAPIKeyCommand{
		ID:        result.Key.ID,
		UserID:    "user1",
		Name:      "new-name",
		ExpiresAt: &expiry,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Name != "new-name" {
		t.Errorf("expected name new-name, got: %s", updated.Name)
	}
	if updated.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestAPIKeyService_ListAPIKeys(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	// Create 3 keys for user1, 1 for user2
	for i := 0; i < 3; i++ {
		_, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "k"})
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user2", Name: "k"})
	if err != nil {
		t.Fatal(err)
	}

	keys, err := svc.ListAPIKeys(ctx, "user1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got: %d", len(keys))
	}
}

func TestAPIKeyService_VerifyAPIKey_Valid(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	result, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "k"})
	if err != nil {
		t.Fatal(err)
	}

	userID, err := svc.VerifyAPIKey(ctx, result.PlainKey)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if userID != "user1" {
		t.Errorf("expected user1, got: %s", userID)
	}
}

func TestAPIKeyService_VerifyAPIKey_Revoked(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)
	ctx := context.Background()

	result, err := svc.CreateAPIKey(ctx, CreateAPIKeyCommand{UserID: "user1", Name: "k"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.RevokeAPIKey(ctx, result.Key.ID, "user1")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := svc.VerifyAPIKey(ctx, result.PlainKey)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if userID != "" {
		t.Errorf("expected empty userID for revoked key, got: %s", userID)
	}
}

func TestAPIKeyService_VerifyAPIKey_NotFound(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := NewAPIKeyService(repo)

	userID, err := svc.VerifyAPIKey(context.Background(), "mc_nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if userID != "" {
		t.Errorf("expected empty userID, got: %s", userID)
	}
}

// ensure errors package is used to satisfy import
var _ = errors.New
