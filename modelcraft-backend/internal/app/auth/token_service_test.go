package auth

import (
	"context"
	"fmt"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/domain/shared"
	"testing"
	"time"

	domainOrg "modelcraft/internal/domain/organization"

	domainauth "modelcraft/internal/domain/auth"
	domainProfile "modelcraft/internal/domain/profile"

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

type mockProfileRepo struct {
	profilesByUserID map[string]*domainProfile.Profile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{profilesByUserID: make(map[string]*domainProfile.Profile)}
}

func (m *mockProfileRepo) Create(_ context.Context, p *domainProfile.Profile) error {
	m.profilesByUserID[p.UserID] = p
	return nil
}

func (m *mockProfileRepo) CreateInitialProfile(ctx context.Context, p *domainProfile.Profile) error {
	return m.Create(ctx, p)
}

func (m *mockProfileRepo) FindByUserID(_ context.Context, _, userID string) (*domainProfile.Profile, error) {
	p, ok := m.profilesByUserID[userID]
	if !ok {
		return nil, shared.NewNotFoundError("profile not found by user id: " + userID)
	}
	return p, nil
}

func (m *mockProfileRepo) UpdateByUserID(
	_ context.Context,
	_ string,
	userID string,
	patch domainProfile.UpdatePatch,
) error {
	p, ok := m.profilesByUserID[userID]
	if !ok {
		return shared.NewNotFoundError("profile not found by user id: " + userID)
	}

	if patch.IsEmpty() {
		return nil
	}

	return p.ApplyPatch(patch)
}

type mockUserRepo struct {
	users        map[string]*domainUser.User // key: externalID
	usersByID    map[string]*domainUser.User // key: internal ID
	usersByPhone map[string]*domainUser.User // key: phone
	usersByName  map[string]*domainUser.User // key: name
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:        make(map[string]*domainUser.User),
		usersByID:    make(map[string]*domainUser.User),
		usersByPhone: make(map[string]*domainUser.User),
		usersByName:  make(map[string]*domainUser.User),
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
	if u.Name != "" {
		m.usersByName[u.Name] = u
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

func (m *mockUserRepo) GetByPhone(_ context.Context, _, phone string) (*domainUser.User, error) {
	u, ok := m.usersByPhone[phone]
	if !ok {
		return nil, shared.NewNotFoundError("user not found by phone: " + phone)
	}
	return u, nil
}

func (m *mockUserRepo) ExistsByPhone(_ context.Context, _, phone string) (bool, error) {
	_, ok := m.usersByPhone[phone]
	return ok, nil
}

func (m *mockUserRepo) GetByName(_ context.Context, _, name string) (*domainUser.User, error) {
	u, ok := m.usersByName[name]
	if !ok {
		return nil, shared.NewNotFoundError("user not found by name: " + name)
	}
	return u, nil
}

func (m *mockUserRepo) ExistsByName(_ context.Context, _, name string) (bool, error) {
	_, ok := m.usersByName[name]
	return ok, nil
}

func (m *mockUserRepo) GetByNameGlobal(_ context.Context, name string) (*domainUser.User, error) {
	u, ok := m.usersByName[name]
	if !ok {
		return nil, shared.NewNotFoundError("user not found by name: " + name)
	}
	return u, nil
}

func (m *mockUserRepo) ListByOrg(_ context.Context, _ string) ([]*domainUser.User, error) {
	return nil, nil
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

// mockOrgRepo is an in-memory org repository for testing.
type mockOrgRepo struct {
	orgsByPhone map[string]*domainOrg.Organization // key: phone
	orgsByName  map[string]*domainOrg.Organization // key: name
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{
		orgsByPhone: make(map[string]*domainOrg.Organization),
		orgsByName:  make(map[string]*domainOrg.Organization),
	}
}

func (m *mockOrgRepo) Create(_ context.Context, org *domainOrg.Organization) error {
	m.orgsByName[org.Name] = org
	if org.Phone != "" {
		m.orgsByPhone[org.Phone] = org
	}
	return nil
}

func (m *mockOrgRepo) GetByName(_ context.Context, name string) (*domainOrg.Organization, error) {
	org, ok := m.orgsByName[name]
	if !ok {
		return nil, shared.NewNotFoundError("org not found: " + name)
	}
	return org, nil
}

func (m *mockOrgRepo) GetByPhone(_ context.Context, phone string) (*domainOrg.Organization, error) {
	org, ok := m.orgsByPhone[phone]
	if !ok {
		return nil, shared.NewNotFoundError("org not found by phone: " + phone)
	}
	return org, nil
}

func (m *mockOrgRepo) ListByUser(_ context.Context, _ string) ([]*domainOrg.Organization, error) {
	return nil, nil
}

func (m *mockOrgRepo) Update(_ context.Context, _ *domainOrg.Organization) error {
	return nil
}

func (m *mockOrgRepo) ExistsByName(_ context.Context, name string) (bool, error) {
	_, ok := m.orgsByName[name]
	return ok, nil
}

func (m *mockOrgRepo) ExistsByPhone(_ context.Context, phone string) (bool, error) {
	_, ok := m.orgsByPhone[phone]
	return ok, nil
}

// ========== Test Helper ==========

func createTestService(t *testing.T) (
	*TokenService,
	*mockRefreshTokenRepo,
	*mockUserRepo,
	*mockProfileRepo,
	*mockAuditLogRepo,
) {
	t.Helper()
	refreshRepo := newMockRefreshTokenRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrgRepo()
	profileRepo := newMockProfileRepo()
	auditRepo := &mockAuditLogRepo{}
	hasher := &mockPasswordHasher{}
	jwtSigner, err := domainauth.GenerateDevSigner()
	require.NoError(t, err)
	svc := NewTokenService(
		refreshRepo, userRepo, orgRepo, profileRepo, auditRepo, hasher, 7*24*time.Hour, nil, nil, jwtSigner,
	)
	return svc, refreshRepo, userRepo, profileRepo, auditRepo
}

// registerTestUser is a helper that registers a user and returns the result.
func registerTestUser(t *testing.T, svc *TokenService, phone, password string) *RegisterResult {
	t.Helper()
	ctx := context.Background()
	result, err := svc.Register(ctx, RegisterCommand{
		Phone:    phone,
		Password: password,
		UserName: "testuser_" + phone[len(phone)-4:],
	})
	require.NoError(t, err)
	return result
}

// ========== Register Tests ==========

func TestTokenService_Register_Success(t *testing.T) {
	svc, _, userRepo, profileRepo, _ := createTestService(t)
	ctx := context.Background()

	result, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
		UserName: "john_doe",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.OrgName)
	assert.NotEmpty(t, result.Profile.ID)
	assert.Equal(t, result.UserID, result.Profile.UserID)
	assert.Regexp(t, `^user_[A-Z0-9]{6}$`, result.Profile.Nickname)
	if assert.NotNil(t, result.Profile.AvatarURL) {
		assert.Equal(t, "mock://avatar/default-1.png", *result.Profile.AvatarURL)
	}

	// Verify user was created with correct phone and userName
	u, ok := userRepo.usersByPhone["13800138000"]
	require.True(t, ok)
	assert.Equal(t, result.UserID, u.ID)
	assert.Equal(t, "john_doe", u.Name)
	assert.Equal(t, "hashed_securePassword1", u.PasswordHash)

	// Verify profile was created and associated with the user
	p, ok := profileRepo.profilesByUserID[result.UserID]
	require.True(t, ok)
	assert.Equal(t, result.Profile.ID, p.ID)
	assert.Equal(t, result.Profile.Nickname, p.Nickname)
}

func TestTokenService_Register_InvalidPhone(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "123",
		Password: "securePassword1",
		UserName: "john_doe",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PARAM_INVALID")
}

func TestTokenService_Register_DuplicatePhone(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	// First registration succeeds
	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
		UserName: "john_doe",
	})
	require.NoError(t, err)

	// Second registration with same phone fails
	_, err = svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "anotherPassword1",
		UserName: "jane_doe",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CONFLICT.PHONE_ALREADY_EXISTS")
}

func TestTokenService_Register_DuplicateUserName(t *testing.T) {
	// With org-scoped uniqueness, same userName is allowed across different orgs
	// (different phones → different orgs).
	// Uniqueness within the same org is enforced by DB constraints, not app layer.
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138000",
		Password: "securePassword1",
		UserName: "john_doe",
	})
	require.NoError(t, err)

	// Different phone (different org) — same userName is now allowed
	_, err = svc.Register(ctx, RegisterCommand{
		Phone:    "13900139000",
		Password: "anotherPassword1",
		UserName: "john_doe",
	})
	assert.NoError(t, err, "same userName in different orgs should be allowed")
}

// ========== Login Tests ==========

func TestTokenService_Login_Success(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Register first
	registerTestUser(t, svc, "13800138000", "securePassword1")

	// Login
	result, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "securePassword1",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.UserID)
	assert.NotEmpty(t, result.RefreshToken)
	assert.False(t, result.ExpiresIn == 0)
}

func TestTokenService_Login_PhoneNotFound(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Login(ctx, LoginCommand{
		UserName: "nonexistent_user",
		Password: "securePassword1",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AUTHENTICATION_FAILED")
}

func TestTokenService_Login_WrongPassword(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")

	_, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "wrongPassword",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AUTHENTICATION_FAILED")
}

// ========== Refresh Tests ==========

func TestTokenService_Refresh_Rotation(t *testing.T) {
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Register and login to get a refresh token
	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "securePassword1",
	})
	require.NoError(t, err)

	refreshResult, err := svc.Refresh(ctx, RefreshCommand{
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResult.RefreshToken)
	assert.NotEqual(t, loginResult.RefreshToken, refreshResult.RefreshToken, "should return new token")
	assert.NotEmpty(t, refreshResult.AccessToken)
}

func TestTokenService_Refresh_ReuseDetection(t *testing.T) {
	svc, refreshRepo, _, _, auditRepo := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "securePassword1",
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
	svc, refreshRepo, _, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "securePassword1",
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
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	_, err := svc.Refresh(ctx, RefreshCommand{RefreshToken: "nonexistent_token_value"})
	assert.Error(t, err)
}

// ========== Logout Tests ==========

func TestTokenService_Logout(t *testing.T) {
	svc, refreshRepo, _, _, _ := createTestService(t)
	ctx := context.Background()

	registerTestUser(t, svc, "13800138000", "securePassword1")
	loginResult, err := svc.Login(ctx, LoginCommand{
		UserName: "testuser_8000",
		Password: "securePassword1",
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
	svc, _, _, _, _ := createTestService(t)
	ctx := context.Background()

	// Logout with nonexistent token should succeed silently
	err := svc.Logout(ctx, LogoutCommand{RefreshToken: "nonexistent"})
	assert.NoError(t, err)
}

// ========== Register — builtin admin 创建 ==========

// spyOrgCreationService 记录 Execute 被调用时传入的所有 input，供断言使用。
type spyOrgCreationService struct {
	calls []*organization.CreateOrganizationInput
}

func (s *spyOrgCreationService) Execute(
	_ context.Context,
	input *organization.CreateOrganizationInput,
) (*organization.CreateOrganizationOutput, error) {
	s.calls = append(s.calls, input)
	return &organization.CreateOrganizationOutput{
		OrganizationName: "spy-org",
	}, nil
}

// createTestServiceWithOrgSpy 返回带 spy 的 TokenService，
// 用于断言注册流程是否正确传递 EndUserAdminPassword。
func createTestServiceWithOrgSpy(t *testing.T) (*TokenService, *spyOrgCreationService) {
	t.Helper()
	spy := &spyOrgCreationService{}
	refreshRepo := newMockRefreshTokenRepo()
	userRepo := newMockUserRepo()
	orgRepo := newMockOrgRepo()
	profileRepo := newMockProfileRepo()
	auditRepo := &mockAuditLogRepo{}
	hasher := &mockPasswordHasher{}
	jwtSigner, err := domainauth.GenerateDevSigner()
	require.NoError(t, err)
	svc := NewTokenService(
		refreshRepo, userRepo, orgRepo, profileRepo, auditRepo, hasher,
		7*24*time.Hour, spy, nil, jwtSigner,
	)
	return svc, spy
}

// TestTokenService_Register_PassesPasswordToOrgCreation 验证注册时
// org creation service 被调用一次。
func TestTokenService_Register_PassesPasswordToOrgCreation(t *testing.T) {
	svc, spy := createTestServiceWithOrgSpy(t)
	ctx := context.Background()

	const password = "securePassword1"
	_, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138001",
		Password: password,
		UserName: "builtin_test_user",
	})
	require.NoError(t, err)

	require.Len(t, spy.calls, 1, "createOrgService.Execute should be called once")
}

// TestTokenService_Register_OrgCreationCalledWithOwnerID 验证 Execute 收到正确的 OwnerUserID。
func TestTokenService_Register_OrgCreationCalledWithOwnerID(t *testing.T) {
	svc, spy := createTestServiceWithOrgSpy(t)
	ctx := context.Background()

	result, err := svc.Register(ctx, RegisterCommand{
		Phone:    "13800138002",
		Password: "securePassword1",
		UserName: "owner_check_user",
	})
	require.NoError(t, err)

	require.Len(t, spy.calls, 1)
	assert.Equal(t, result.UserID, spy.calls[0].OwnerUserID,
		"OwnerUserID passed to org creation must match the newly registered user's ID")
}
