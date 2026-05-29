package enduser

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
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
	userRepo    *inMemoryEndUserRepo
	sessionRepo *inMemoryEndUserSessionRepo
}

func (f *fakeRepoFactory) NewEndUserRepository(
	_ SQLDBTX,
	_, _ string,
) domainenduser.EndUserRepository {
	return f.userRepo
}

func (f *fakeRepoFactory) NewEndUserSessionRepository(
	_ SQLDBTX,
	_, _ string,
) domainenduser.EndUserSessionRepository {
	return f.sessionRepo
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
	_ string,
	_ string,
) (*domainenduser.EndUser, error) {
	return nil, nil //nolint:nilnil
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

type inMemoryEndUserSessionRepo struct {
	sessionsByID   map[string]*domainenduser.EndUserSession
	sessionsByHash map[string]*domainenduser.EndUserSession
	forceSaveErr   error
	forceFindErr   error
	forceRevokeErr error
}

func newInMemoryEndUserSessionRepo() *inMemoryEndUserSessionRepo {
	return &inMemoryEndUserSessionRepo{
		sessionsByID:   make(map[string]*domainenduser.EndUserSession),
		sessionsByHash: make(map[string]*domainenduser.EndUserSession),
	}
}

func (r *inMemoryEndUserSessionRepo) Save(
	_ context.Context,
	session *domainenduser.EndUserSession,
) error {
	if r.forceSaveErr != nil {
		return r.forceSaveErr
	}
	s := *session
	r.sessionsByID[s.ID] = &s
	r.sessionsByHash[s.RefreshTokenHash] = &s
	return nil
}

func (r *inMemoryEndUserSessionRepo) GetByTokenHash(
	_ context.Context,
	tokenHash string,
) (*domainenduser.EndUserSession, error) {
	if r.forceFindErr != nil {
		return nil, r.forceFindErr
	}
	session, ok := r.sessionsByHash[tokenHash]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (r *inMemoryEndUserSessionRepo) RevokeByID(
	_ context.Context,
	id string,
) error {
	if r.forceRevokeErr != nil {
		return r.forceRevokeErr
	}
	session, ok := r.sessionsByID[id]
	if !ok {
		return nil
	}
	session.Revoked = true
	return nil
}

func (r *inMemoryEndUserSessionRepo) RevokeAllByUserID(
	_ context.Context,
	userID string,
) error {
	for _, session := range r.sessionsByID {
		if session.UserID == userID {
			session.Revoked = true
		}
	}
	return nil
}

func createEndUserAuthServiceForTest(t *testing.T) (
	*EndUserAuthAppService,
	*inMemoryEndUserRepo,
	*inMemoryEndUserSessionRepo,
) {
	t.Helper()

	userRepo := newInMemoryEndUserRepo()
	sessionRepo := newInMemoryEndUserSessionRepo()

	svc := NewEndUserAuthAppService(
		&sql.DB{},
		&fakeRepoFactory{userRepo: userRepo, sessionRepo: sessionRepo},
		&fakeTxManager{},
		&fakeTokenIssuer{},
		logfacade.GetLogger(context.Background()),
	)

	return svc, userRepo, sessionRepo
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
	svc, userRepo, sessionRepo := createEndUserAuthServiceForTest(t)
	seedEndUser(t, userRepo, "org-a", "user-1", "alice", "Password123", false)
	userRepo.accessibleProjects["user-1"] = []domainenduser.AccessibleProject{
		{ProjectSlug: "project-a", ProjectTitle: "Project A"},
		{ProjectSlug: "project-b", ProjectTitle: "Project B"},
	}

	result, err := svc.LoginEndUser(context.Background(), LoginCommand{
		OrgName:  "org-a",
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-1", result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.Len(t, result.Projects, 2)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Len(t, sessionRepo.sessionsByID, 1)

	issuer, ok := svc.tokenIssuer.(*fakeTokenIssuer)
	require.True(t, ok)
	require.Len(t, issuer.issuedInputs, 1)
	assert.Equal(t, "user-1", issuer.issuedInputs[0].UserID)
	assert.Equal(t, "org-a", issuer.issuedInputs[0].OrgName)
	assert.Equal(t, []string{"project-a", "project-b"}, issuer.issuedInputs[0].ProjectSlugs)
}

func TestEndUserAuthAppService_LoginEndUser_ResolveOrgFromUsername(t *testing.T) {
	svc, userRepo, sessionRepo := createEndUserAuthServiceForTest(t)
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
	assert.Len(t, sessionRepo.sessionsByID, 1)

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
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.Projects)
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
	svc, userRepo, sessionRepo := createEndUserAuthServiceForTest(t)
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

	refreshResult, err := svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	require.NotNil(t, refreshResult)

	assert.NotEqual(t, loginResult.RefreshToken, refreshResult.RefreshToken)
	assert.Equal(t, "user-1", refreshResult.UserID)
	assert.NotEmpty(t, refreshResult.AccessToken)
	assert.Len(t, refreshResult.Projects, 1)

	oldHash := hashToken(loginResult.RefreshToken)
	oldSession := sessionRepo.sessionsByHash[oldHash]
	require.NotNil(t, oldSession)
	assert.True(t, oldSession.Revoked)
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

func TestEndUserAuthAppService_RefreshEndUserToken_NoProjectAccess(t *testing.T) {
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

	userRepo.accessibleProjects["user-1"] = nil

	refreshResult, err := svc.RefreshEndUserToken(context.Background(), RefreshCommand{
		OrgName:      "org-a",
		RefreshToken: loginResult.RefreshToken,
	})
	require.NoError(t, err)
	require.NotNil(t, refreshResult)
	assert.Equal(t, "user-1", refreshResult.UserID)
	assert.Empty(t, refreshResult.AccessToken)
	assert.Empty(t, refreshResult.Projects)
	assert.NotEmpty(t, refreshResult.RefreshToken)
}

var (
	_ driver.Result                          = noopResult{}
	_ SQLDBTX                                = noopSQLDBTX{}
	_ domainenduser.EndUserRepository        = (*inMemoryEndUserRepo)(nil)
	_ domainenduser.EndUserSessionRepository = (*inMemoryEndUserSessionRepo)(nil)
)
