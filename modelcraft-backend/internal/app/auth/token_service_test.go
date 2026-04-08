package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	domainauth "modelcraft/internal/domain/auth"
	"modelcraft/internal/domain/shared"
	domainUser "modelcraft/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Mock Repositories ==========

type mockRefreshTokenRepo struct {
	tokens map[string]*domainauth.RefreshToken // key: tokenHash
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{tokens: make(map[string]*domainauth.RefreshToken)}
}

func (m *mockRefreshTokenRepo) Save(_ context.Context, token *domainauth.RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *mockRefreshTokenRepo) FindByHash(_ context.Context, hash string) (*domainauth.RefreshToken, error) {
	t, ok := m.tokens[hash]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockRefreshTokenRepo) Revoke(_ context.Context, id string) error {
	for _, t := range m.tokens {
		if t.ID == id {
			now := time.Now()
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(_ context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			now := time.Now()
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(_ context.Context) error { return nil }

type mockAuditLogRepo struct {
	logs []*domainauth.SecurityAuditLog
}

func (m *mockAuditLogRepo) Insert(_ context.Context, log *domainauth.SecurityAuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

type mockUserRepo struct {
	users       map[string]*domainUser.User // key: externalID
	usersByID   map[string]*domainUser.User // key: internal ID
	usersByPhone map[string]*domainUser.User // key: phone
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:        make(map[string]*domainUser.User),
		usersByID:    make(map[string]*domainUser.User),
		usersByPhone: make(map[string]*domainUser.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, u *domainUser.User) error {
	if u.ExternalID != "" {
		m.users[u.ExternalID] = u
	}
	m.usersByID[u.ID] = u
	if !u.Phone.IsZero() {
		m.usersByPhone[u.Phone.String()] = u
	}
	return nil
}

func (m *mockUserRepo) GetByExternalID(_ context.Context, externalID string) (*domainUser.User, error) {
	u, ok := m.users[externalID]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domainUser.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) ExistsByExternalID(_ context.Context, externalID string) (bool, error) {
	_, ok := m.users[externalID]
	return ok, nil
}

func (m *mockUserRepo) FindIDByExternalID(_ context.Context, externalID string) (string, bool, error) {
	if u, ok := m.users[externalID]; ok {
		return u.ID, true, nil
	}
	return "", false, nil
}

func (m *mockUserRepo) GetByPhone(_ context.Context, phone string) (*domainUser.User, error) {
	u, ok := m.usersByPhone[phone]
	if !ok {
		return nil, shared.NewNotFoundError("user not found by phone: " + phone)
	}
	return u, nil
}

func (m *mockUserRepo) ExistsByPhone(_ context.Context, phone string) (bool, error) {
	_, ok := m.usersByPhone[phone]
	return ok, nil
}

// mockPasswordHasher is an in-memory password hasher for testing.
type mockPasswordHasher struct{}

func (m *mockPasswordHasher) Hash(_ context.Context, password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *mockPasswordHasher) Verify(_ context.Context, password, hash string) error {
	if hash == "hashed_"+password {
		return nil
	}
	return fmt.Errorf("password mismatch")
}

// ========== Test Helper ==========

func createTestService(t *testing.T) (*TokenService, *mockRefreshTokenRepo, *mockUserRepo, *mockAuditLogRepo) {
	t.Helper()
	refreshRepo := newMockRefreshTokenRepo()
	userRepo := newMockUserRepo()
	auditRepo := &mockAuditLogRepo{}
	hasher := &mockPasswordHasher{}
	svc := NewTokenService(refreshRepo, userRepo, auditRepo, hasher, 7*24*time.Hour)
	return svc, refreshRepo, userRepo, auditRepo
}

// registerTestUser is a helper that registers a user and returns the result.
func registerTestUser(t *testing.T, svc *TokenService, phone, password string) *RegisterResult {
	t.Helper()
	ctx := context.Background()
	result, err := svc.Register(ctx, RegisterCommand{Phone: phone, Password: password})
	require.NoError(t, err)
	return result
}

// ========== Register Tests ==========

func TestTokenService_Register_Success(t *testing.T) {
	svc, _, userRepo, _ := createTestService(t)
	ctx := context.Background()

	result, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)

	// Verify user was created with correct phone
	u, ok := userRepo.usersByPhone["13800138000"]
	require.True(t, ok)
	assert.Equal(t, result.UserID, u.ID)
	assert.Equal(t, "138****8000", u.Name) // auto-masked name
	assert.Equal(t, "hashed_securePassword1", u.PasswordHash)
}

func TestTokenService_Register_InvalidPhone(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "123",
		Password: "securePassword1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PARAM_INVALID")
}

func TestTokenService_Register_WeakPassword(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "short",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PARAM_INVALID")
}

func TestTokenService_Register_DuplicatePhone(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	// First registration succeeds
	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
	})
	require.NoError(t, err)

	// Second registration with same phone fails
	_, err = svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "anotherPassword1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFLICT")
}

// ========== Login Tests ==========

func TestTokenService_Login_Success(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Register first
	registerTestUser(t, svc, "13800138000", "securePassword1")

	// Login
	result, err := svc.Login(ctx, LoginCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.RefreshToken)
	assert.False(t, result.ExpiresAt.Before(time.Now()))
}

func TestTokenService_Login_InvalidPhone(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Login(ctx, LoginCommand{
		Phone:    "invalid",
		Password: "securePassword1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PARAM_INVALID")
}

func TestTokenService_Login_PhoneNotFound(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Login(ctx, LoginCommand{
		Phone:    "13900139000",
		Password: "securePassword1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AUTHENTICATION_FAILED")
}

func TestTokenService_Login_WrongPassword(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")

	_, err := svc.Login(ctx, LoginCommand{
		Phone:    "13800138000",
		Password: "wrongPassword",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AUTHENTICATION_FAILED")
}

// ========== OAuth Login Tests (backward compat) ==========

func TestTokenService_OAuthLogin_NewUser(t *testing.T) {
	svc, _, userRepo, _ := createTestService(t)
	ctx := context.Background()

	result, err := svc.OAuthLogin(ctx, OAuthLoginCommand{
		ExternalID: "ext_123",
		Email:      "test@example.com",
		Name:       "Test User",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.RefreshToken)
	assert.False(t, result.ExpiresAt.Before(time.Now()))

	// Verify user was created
	u, err := userRepo.GetByExternalID(ctx, "ext_123")
	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "Test User", u.Name)
}

func TestTokenService_OAuthLogin_ExistingUser(t *testing.T) {
	svc, _, userRepo, _ := createTestService(t)
	ctx := context.Background()

	// Pre-create user
	existingUser, _ := domainUser.NewOAuthUser("pre-existing-id", "ext_456", "Existing", "")
	_ = userRepo.Create(ctx, existingUser)

	result, err := svc.OAuthLogin(ctx, OAuthLoginCommand{
		ExternalID: "ext_456",
		Email:      "updated@example.com",
		Name:       "Existing",
	})

	require.NoError(t, err)
	assert.Equal(t, "pre-existing-id", result.UserID)
}

// ========== Refresh Tests ==========

func TestTokenService_Refresh_Rotation(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Register and login to get a refresh token
	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		Phone: "13800138000", Password: "securePassword1",
	})
	require.NoError(t, err)

	refreshResult, err := svc.Refresh(ctx, RefreshCommand{
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResult.RefreshToken)
	assert.NotEqual(t, loginResult.RefreshToken, refreshResult.RefreshToken, "should return new token")
	assert.Equal(t, loginResult.UserID, refreshResult.UserID)
}

func TestTokenService_Refresh_ReuseDetection(t *testing.T) {
	svc, refreshRepo, _, auditRepo := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		Phone: "13800138000", Password: "securePassword1",
	})
	require.NoError(t, err)

	// First refresh consumes the original token
	_, err = svc.Refresh(ctx, RefreshCommand{RefreshToken: loginResult.RefreshToken})
	require.NoError(t, err)

	// Second refresh with the same (now revoked) token triggers reuse detection
	_, err = svc.Refresh(ctx, RefreshCommand{RefreshToken: loginResult.RefreshToken})
	assert.Error(t, err)

	// Verify audit log
	require.Len(t, auditRepo.logs, 1)
	assert.Equal(t, domainauth.EventReuseDetected, auditRepo.logs[0].Event)

	// Verify all user tokens were revoked
	for _, token := range refreshRepo.tokens {
		if token.UserID == loginResult.UserID {
			assert.True(t, token.IsRevoked())
		}
	}
}

func TestTokenService_Refresh_ExpiredToken(t *testing.T) {
	svc, refreshRepo, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		Phone: "13800138000", Password: "securePassword1",
	})
	require.NoError(t, err)

	// Manually expire the token
	hash := HashToken(loginResult.RefreshToken)
	token := refreshRepo.tokens[hash]
	token.ExpiresAt = time.Now().Add(-time.Hour)

	_, err = svc.Refresh(ctx, RefreshCommand{RefreshToken: loginResult.RefreshToken})
	assert.Error(t, err)
}

func TestTokenService_Refresh_UnknownToken(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Refresh(ctx, RefreshCommand{RefreshToken: "nonexistent_token_value"})
	assert.Error(t, err)
}

// ========== Logout Tests ==========

func TestTokenService_Logout(t *testing.T) {
	svc, refreshRepo, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		Phone: "13800138000", Password: "securePassword1",
	})
	require.NoError(t, err)

	err = svc.Logout(ctx, LogoutCommand{RefreshToken: loginResult.RefreshToken})
	require.NoError(t, err)

	// Verify token is revoked
	hash := HashToken(loginResult.RefreshToken)
	token := refreshRepo.tokens[hash]
	assert.True(t, token.IsRevoked())

	// Refresh after logout should fail
	_, err = svc.Refresh(ctx, RefreshCommand{RefreshToken: loginResult.RefreshToken})
	assert.Error(t, err)
}

func TestTokenService_Logout_NonexistentToken(t *testing.T) {
	svc, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Logout with nonexistent token should succeed silently
	err := svc.Logout(ctx, LogoutCommand{RefreshToken: "nonexistent"})
	assert.NoError(t, err)
}
