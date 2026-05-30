package enduser

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	domainAuth "modelcraft/internal/domain/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainenduser "modelcraft/internal/domain/enduser"
)

type fakeTokenIssuer struct {
	forceErr     error
	issuedInputs []EndUserTokenIssueInput
}

func (i *fakeTokenIssuer) IssueEndUserToken(
	_ context.Context,
	input EndUserTokenIssueInput,
) (*EndUserTokenIssueResult, error) {
	if i.forceErr != nil {
		return nil, i.forceErr
	}

	recordedInput := input
	recordedInput.ProjectSlugs = append([]string(nil), input.ProjectSlugs...)
	i.issuedInputs = append(i.issuedInputs, recordedInput)

	return &EndUserTokenIssueResult{
		AccessToken: "token-for-" + input.UserID,
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil
}

type fakeRepoFactory struct {
	userRepo         *inMemoryEndUserRepo
	refreshTokenRepo *inMemoryRefreshTokenRepo
}

func (f *fakeRepoFactory) NewEndUserRepository(
	_ SQLDBTX,
	_, _ string,
) domainenduser.EndUserRepository {
	return f.userRepo
}

func (f *fakeRepoFactory) NewRefreshTokenRepository(
	_ SQLDBTX,
) domainAuth.RefreshTokenRepository {
	return f.refreshTokenRepo
}

type fakeTxManager struct{}

func (m *fakeTxManager) WithTx(
	ctx context.Context,
	_ *sql.DB,
	fn func(ctx context.Context, txDB SQLDBTX) error,
) error {
	return fn(ctx, noopSQLDBTX{})
}

type noopSQLDBTX struct{}

type noopResult struct{}

func (noopResult) LastInsertId() (int64, error) { return 0, nil }
func (noopResult) RowsAffected() (int64, error) { return 0, nil }

func (noopSQLDBTX) PrepareContext(
	_ context.Context,
	_ string,
) (*sql.Stmt, error) {
	panic("PrepareContext not implemented in noopSQLDBTX")
}

func (noopSQLDBTX) ExecContext(
	_ context.Context,
	_ string,
	_ ...any,
) (sql.Result, error) {
	return noopResult{}, nil
}

func (noopSQLDBTX) QueryContext(
	_ context.Context,
	_ string,
	_ ...any,
) (*sql.Rows, error) {
	return nil, errors.New("not implemented in noop SQLDBTX")
}

func (noopSQLDBTX) QueryRowContext(
	_ context.Context,
	_ string,
	_ ...any,
) *sql.Row {
	return &sql.Row{}
}

type inMemoryEndUserRepo struct {
	usersByID          map[string]*domainenduser.EndUser
	usersByOrgName     map[string]map[string]*domainenduser.EndUser
	usersByPhone       map[string]map[string]*domainenduser.EndUser // orgName → phone → user
	forceSaveErr       error
	forceLookupErr     error
	forceDeleteErr     error
	forceUpdateErr     error
	forceListErr       error
	forceGetByIDErr    error
	accessibleProjects map[string][]domainenduser.AccessibleProject
}

func newInMemoryEndUserRepo() *inMemoryEndUserRepo {
	return &inMemoryEndUserRepo{
		usersByID:          make(map[string]*domainenduser.EndUser),
		usersByOrgName:     make(map[string]map[string]*domainenduser.EndUser),
		usersByPhone:       make(map[string]map[string]*domainenduser.EndUser),
		accessibleProjects: make(map[string][]domainenduser.AccessibleProject),
	}
}

func (r *inMemoryEndUserRepo) Save(
	_ context.Context,
	user *domainenduser.EndUser,
) error {
	if r.forceSaveErr != nil {
		return r.forceSaveErr
	}
	if _, ok := r.usersByOrgName[user.OrgName]; !ok {
		r.usersByOrgName[user.OrgName] = make(map[string]*domainenduser.EndUser)
	}
	r.usersByID[user.ID] = user
	r.usersByOrgName[user.OrgName][user.Username] = user
	return nil
}

func (r *inMemoryEndUserRepo) GetByID(
	_ context.Context,
	orgName,
	id string,
) (*domainenduser.EndUser, error) {
	if r.forceGetByIDErr != nil {
		return nil, r.forceGetByIDErr
	}
	user, ok := r.usersByID[id]
	if !ok || user.OrgName != orgName {
		return nil, nil
	}
	return user, nil
}

func (r *inMemoryEndUserRepo) GetByUsername(
	_ context.Context,
	orgName,
	username string,
) (*domainenduser.EndUser, error) {
	if r.forceLookupErr != nil {
		return nil, r.forceLookupErr
	}
	usersInOrg, ok := r.usersByOrgName[orgName]
	if !ok {
		return nil, nil
	}
	user, ok := usersInOrg[username]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *inMemoryEndUserRepo) GetByPhone(
	_ context.Context,
	orgName string,
	phone string,
) (*domainenduser.EndUser, error) {
	if r.forceLookupErr != nil {
		return nil, r.forceLookupErr
	}
	usersInOrg, ok := r.usersByPhone[orgName]
	if !ok {
		return nil, nil //nolint:nilnil
	}
	user, ok := usersInOrg[phone]
	if !ok {
		return nil, nil //nolint:nilnil
	}
	return user, nil
}

func (r *inMemoryEndUserRepo) GetByUsernameGlobal(
	_ context.Context,
	username string,
) (*domainenduser.EndUser, error) {
	if r.forceLookupErr != nil {
		return nil, r.forceLookupErr
	}
	for _, user := range r.usersByID {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (r *inMemoryEndUserRepo) UpdateStatus(
	_ context.Context,
	orgName,
	id string,
	isForbidden bool,
) error {
	if r.forceUpdateErr != nil {
		return r.forceUpdateErr
	}
	user, ok := r.usersByID[id]
	if !ok || user.OrgName != orgName {
		return nil
	}
	user.IsForbidden = isForbidden
	return nil
}

func (r *inMemoryEndUserRepo) Delete(
	_ context.Context,
	orgName,
	id string,
) error {
	if r.forceDeleteErr != nil {
		return r.forceDeleteErr
	}
	user, ok := r.usersByID[id]
	if !ok || user.OrgName != orgName {
		return nil
	}
	delete(r.usersByID, id)
	delete(r.usersByOrgName[orgName], user.Username)
	return nil
}

func (r *inMemoryEndUserRepo) ListWithTotal(
	_ context.Context,
	_ domainenduser.ListEndUsersQuery,
) ([]*domainenduser.EndUser, int64, error) {
	if r.forceListErr != nil {
		return nil, 0, r.forceListErr
	}
	return []*domainenduser.EndUser{}, 0, nil
}

func (r *inMemoryEndUserRepo) ListAccessibleProjectsByRoleAssignment(
	_ context.Context,
	_, endUserID string,
) ([]domainenduser.AccessibleProject, error) {
	return r.accessibleProjects[endUserID], nil
}

func (r *inMemoryEndUserRepo) HasProjectAccessByRole(
	_ context.Context,
	_, endUserID, projectSlug string,
) (bool, error) {
	for _, p := range r.accessibleProjects[endUserID] {
		if p.ProjectSlug == projectSlug {
			return true, nil
		}
	}
	return false, nil
}

func (r *inMemoryEndUserRepo) ListAllProjectsByOrg(
	_ context.Context,
	_ string,
) ([]domainenduser.AccessibleProject, error) {
	return []domainenduser.AccessibleProject{}, nil
}

func (r *inMemoryEndUserRepo) GetBuiltinByOrg(_ context.Context, _ string) (*domainenduser.EndUser, error) {
	return nil, nil //nolint:nilnil
}

func (r *inMemoryEndUserRepo) UpdatePassword(
	_ context.Context, _, id string, hashedPassword domainenduser.HashedPassword,
) error {
	u, ok := r.usersByID[id]
	if !ok {
		return fmt.Errorf("user not found: %s", id)
	}
	u.Password = hashedPassword
	return nil
}

// inMemoryRefreshTokenRepo implements domainAuth.RefreshTokenRepository for tests.
type inMemoryRefreshTokenRepo struct {
	tokensByID     map[string]*domainAuth.RefreshToken
	tokensByHash   map[string]*domainAuth.RefreshToken
	forceSaveErr   error
	forceFindErr   error
	forceRevokeErr error
}

func newInMemoryRefreshTokenRepo() *inMemoryRefreshTokenRepo {
	return &inMemoryRefreshTokenRepo{
		tokensByID:   make(map[string]*domainAuth.RefreshToken),
		tokensByHash: make(map[string]*domainAuth.RefreshToken),
	}
}

func (r *inMemoryRefreshTokenRepo) Save(
	_ context.Context,
	token *domainAuth.RefreshToken,
) error {
	if r.forceSaveErr != nil {
		return r.forceSaveErr
	}
	t := *token
	r.tokensByID[t.ID] = &t
	r.tokensByHash[t.TokenHash] = &t
	return nil
}

func (r *inMemoryRefreshTokenRepo) FindByHash(
	_ context.Context,
	hash string,
) (*domainAuth.RefreshToken, error) {
	if r.forceFindErr != nil {
		return nil, r.forceFindErr
	}
	token, ok := r.tokensByHash[hash]
	if !ok {
		return nil, nil //nolint:nilnil
	}
	return token, nil
}

func (r *inMemoryRefreshTokenRepo) Revoke(
	_ context.Context,
	id string,
) error {
	if r.forceRevokeErr != nil {
		return r.forceRevokeErr
	}
	token, ok := r.tokensByID[id]
	if !ok {
		return nil
	}
	now := time.Now()
	token.RevokedAt = &now
	return nil
}

func (r *inMemoryRefreshTokenRepo) RevokeAllByUserID(
	_ context.Context,
	userID string,
) error {
	now := time.Now()
	for _, token := range r.tokensByID {
		if token.UserID == userID {
			token.RevokedAt = &now
		}
	}
	return nil
}

func (r *inMemoryRefreshTokenRepo) DeleteExpired(_ context.Context) error {
	return nil
}

func createEndUserAuthServiceForTest(t *testing.T) (
	*EndUserAuthAppService,
	*inMemoryEndUserRepo,
	*inMemoryRefreshTokenRepo,
) {
	t.Helper()

	userRepo := newInMemoryEndUserRepo()
	refreshTokenRepo := newInMemoryRefreshTokenRepo()

	svc := NewEndUserAuthAppService(
		&sql.DB{},
		&fakeRepoFactory{userRepo: userRepo, refreshTokenRepo: refreshTokenRepo},
		&fakeTxManager{},
		&fakeTokenIssuer{},
		logfacade.GetLogger(context.Background()),
		nil, // auditLogRepo — optional, not needed in unit tests
	)

	return svc, userRepo, refreshTokenRepo
}

func seedEndUser(
	t *testing.T,
	repo *inMemoryEndUserRepo,
	orgName,
	userID,
	username,
	plainPassword string,
	disabled bool,
) {
	t.Helper()

	hashed, err := domainenduser.NewHashedPasswordFromPlain(plainPassword)
	require.NoError(t, err)

	user, err := domainenduser.NewEndUser(userID, orgName, username, hashed)
	require.NoError(t, err)
	if disabled {
		user.Disable()
	}
	require.NoError(t, repo.Save(context.Background(), user))
}

func requireBusinessErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, code, bizErr.Info().GetCode())
}

func TestEndUserAuthAppService_LoginEndUser_Success(t *testing.T) {
	svc, userRepo, refreshTokenRepo := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)

	result, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-1", result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Len(t, refreshTokenRepo.tokensByID, 1)

	issuer, ok := svc.tokenIssuer.(*fakeTokenIssuer)
	require.True(t, ok)
	require.Len(t, issuer.issuedInputs, 1)
	assert.Equal(t, "user-1", issuer.issuedInputs[0].UserID)
	assert.Equal(t, "org-a", issuer.issuedInputs[0].OrgName)
	assert.Nil(t, issuer.issuedInputs[0].ProjectSlugs)
}

func TestEndUserAuthAppService_LoginEndUser_ResolveOrgFromUsername(t *testing.T) {
	svc, userRepo, refreshTokenRepo := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "Project A"},
	}

	result, err := svc.LoginEndUser(context.Background(), LoginCommand{
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "org-a", result.OrgName)
	assert.Equal(t, "user-1", result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Len(t, refreshTokenRepo.tokensByID, 1)

	issuer, ok := svc.tokenIssuer.(*fakeTokenIssuer)
	require.True(t, ok)
	require.Len(t, issuer.issuedInputs, 1)
	assert.Equal(t, "org-a", issuer.issuedInputs[0].OrgName)
}

func TestEndUserAuthAppService_LoginEndUser_NoProjectAccess(t *testing.T) {
	svc, userRepo, _ := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)

	result, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-1", result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestEndUserAuthAppService_LoginEndUser_DisabledAccount(t *testing.T) {
	svc, userRepo, _ := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", true)
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "A"},
	}

	_, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserAccountDisabled.GetCode())
}

func TestEndUserAuthAppService_RefreshEndUserToken_Rotation(t *testing.T) {
	svc, userRepo, refreshTokenRepo := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)

	loginResult, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)

	refreshResult, err := svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	require.NotNil(t, refreshResult)

	assert.NotEqual(t, loginResult.RefreshToken, refreshResult.RefreshToken)
	assert.Equal(t, "user-1", refreshResult.UserID)
	assert.NotEmpty(t, refreshResult.AccessToken)

	oldHash := hashToken(loginResult.RefreshToken)
	oldToken := refreshTokenRepo.tokensByHash[oldHash]
	require.NotNil(t, oldToken)
	assert.True(t, oldToken.IsRevoked())
}

func TestEndUserAuthAppService_RefreshEndUserToken_InvalidToken(t *testing.T) {
	svc, _, _ := createEndUserAuthServiceForTest(t)

	_, err := svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: "invalid-token",
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserInvalidRefreshToken.GetCode())
}

func TestEndUserAuthAppService_RefreshEndUserToken_DisabledAccount(t *testing.T) {
	svc, userRepo, _ := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "A"},
	}

	loginResult, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)

	user, getErr := userRepo.GetByID(context.Background(), "org-a", "user-1")
	require.NoError(t, getErr)
	require.NotNil(t, user)
	user.Disable()
	require.NoError(t, userRepo.UpdateStatus(context.Background(), "org-a", "user-1", true))

	_, err = svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: loginResult.RefreshToken,
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserAccountDisabled.GetCode())
}

func TestEndUserAuthAppService_RefreshEndUserToken_RevokedTokenReuse(t *testing.T) {
	svc, userRepo, _ := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "A"},
	}

	loginResult, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)

	oldToken := loginResult.RefreshToken

	// First refresh: old token → new token (rotation).
	_, err = svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: oldToken,
	})
	require.NoError(t, err)

	// Second refresh: reusing the now-revoked old token must fail.
	_, err = svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: oldToken,
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserInvalidRefreshToken.GetCode())
}

func TestEndUserAuthAppService_RefreshEndUserToken_NoProjectAccess(t *testing.T) {
	svc, userRepo, _ := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)

	loginResult, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)

	refreshResult, err := svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	require.NotNil(t, refreshResult)
	assert.Equal(t, "user-1", refreshResult.UserID)
	assert.NotEmpty(t, refreshResult.AccessToken)
	assert.NotEmpty(t, refreshResult.RefreshToken)
}

func seedEndUserByPhone(
	t *testing.T,
	repo *inMemoryEndUserRepo,
	orgName,
	userID,
	username,
	phone,
	plainPassword string,
) {
	t.Helper()

	hashed, err := domainenduser.NewHashedPasswordFromPlain(plainPassword)
	require.NoError(t, err)

	user, err := domainenduser.NewEndUser(userID, orgName, username, hashed)
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), user))

	if _, ok := repo.usersByPhone[orgName]; !ok {
		repo.usersByPhone[orgName] = make(map[string]*domainenduser.EndUser)
	}
	repo.usersByPhone[orgName][phone] = user
}

func TestEndUserAuthAppService_LoginEndUser_Phone_Success(t *testing.T) {
	svc, userRepo, refreshTokenRepo := createEndUserAuthServiceForTest(t)
	seedEndUserByPhone(t, userRepo, "org-a", "user-1", "alice", "+8613812345678", "Password123")
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "Project A"},
	}

	result, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:        "org-a",
		Identifier:     "+8613812345678",
		IdentifierType: IdentifierTypePhone,
		Password:       "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, "org-a", result.OrgName)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Len(t, refreshTokenRepo.tokensByID, 1)
}

func TestEndUserAuthAppService_LoginEndUser_Phone_EmptyOrgName(t *testing.T) {
	svc, _, _ := createEndUserAuthServiceForTest(t)

	_, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:        "",
		Identifier:     "+8613812345678",
		IdentifierType: IdentifierTypePhone,
		Password:       "Password123",
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserParamInvalid.GetCode())
}

func TestEndUserAuthAppService_LoginEndUser_Phone_EmptyIdentifier(t *testing.T) {
	svc, _, _ := createEndUserAuthServiceForTest(t)

	_, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:        "org-a",
		Identifier:     "",
		IdentifierType: IdentifierTypePhone,
		Password:       "Password123",
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserParamInvalid.GetCode())
}

func TestEndUserAuthAppService_LoginEndUser_Phone_UserNotFound(t *testing.T) {
	svc, _, _ := createEndUserAuthServiceForTest(t)

	_, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:        "org-a",
		Identifier:     "+8613800000000",
		IdentifierType: IdentifierTypePhone,
		Password:       "Password123",
	})
	require.Error(t, err)
	requireBusinessErrorCode(t, err, bizerrors.EndUserInvalidCredentials.GetCode())
}

var (
	_ driver.Result                     = noopResult{}
	_ SQLDBTX                           = noopSQLDBTX{}
	_ domainenduser.EndUserRepository   = (*inMemoryEndUserRepo)(nil)
	_ domainAuth.RefreshTokenRepository = (*inMemoryRefreshTokenRepo)(nil)
)
