package organization_test

import (
	"context"
	"fmt"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/domain/membership"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/domain/user"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/pkg/bizerrors"
	"testing"
	"time"

	domainOrg "modelcraft/internal/domain/organization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock TxManager ---

// mockTxManager is a mock TxManager that returns a controlled error without
// executing the transaction function. Transaction correctness is covered by
// integration tests (task auto-test).
type mockTxManager struct {
	mock.Mock
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// --- Mock Repos ---

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) GetByExternalID(ctx context.Context, externalID string) (*user.User, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	args := m.Called(ctx, externalID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepo) FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error) {
	args := m.Called(ctx, externalID)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*user.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) GetByName(ctx context.Context, name string) (*user.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *mockUserRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	args := m.Called(ctx, phone)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

type mockOrgRepo struct{ mock.Mock }

func (m *mockOrgRepo) Create(ctx context.Context, org *domainOrg.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgRepo) GetByName(ctx context.Context, name string) (*domainOrg.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainOrg.Organization), args.Error(1)
}

func (m *mockOrgRepo) ListByUser(ctx context.Context, userID string) ([]*domainOrg.Organization, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainOrg.Organization), args.Error(1)
}

func (m *mockOrgRepo) Update(ctx context.Context, org *domainOrg.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) CreateRole(ctx context.Context, role *permission.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepo) GetRoleByID(ctx context.Context, id int) (*permission.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Role), args.Error(1)
}

func (m *mockRoleRepo) GetRoleByNameAndOrg(ctx context.Context, name, orgName string) (*permission.Role, error) {
	args := m.Called(ctx, name, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*permission.Role), args.Error(1)
}

func (m *mockRoleRepo) ListRolesByOrg(
	ctx context.Context, orgName string, includeSystem bool,
) ([]*permission.Role, error) {
	args := m.Called(ctx, orgName, includeSystem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*permission.Role), args.Error(1)
}

func (m *mockRoleRepo) UpdateRole(ctx context.Context, role *permission.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepo) DeleteRole(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

type mockMembershipRepo struct{ mock.Mock }

func (m *mockMembershipRepo) Create(ctx context.Context, ms *membership.Membership) error {
	return m.Called(ctx, ms).Error(0)
}

func (m *mockMembershipRepo) GetByID(ctx context.Context, id string) (*membership.Membership, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*membership.Membership), args.Error(1)
}

func (m *mockMembershipRepo) GetByUserAndOrg(
	ctx context.Context, userID, orgID string,
) (*membership.Membership, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*membership.Membership), args.Error(1)
}

func (m *mockMembershipRepo) ListByOrg(ctx context.Context, orgID string) ([]*membership.Membership, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*membership.Membership), args.Error(1)
}

func (m *mockMembershipRepo) ListByOrgWithUserName(
	ctx context.Context, orgID string,
) ([]*membership.MembershipWithUserName, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*membership.MembershipWithUserName), args.Error(1)
}

func (m *mockMembershipRepo) ListByUser(ctx context.Context, userID string) ([]*membership.Membership, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*membership.Membership), args.Error(1)
}

func (m *mockMembershipRepo) ListByUserWithDetails(
	ctx context.Context, userID string, limit int,
) ([]*membership.MembershipWithDetails, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*membership.MembershipWithDetails), args.Error(1)
}

func (m *mockMembershipRepo) CountByUser(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockMembershipRepo) Update(ctx context.Context, ms *membership.Membership) error {
	return m.Called(ctx, ms).Error(0)
}

func (m *mockMembershipRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// --- Helpers ---

func newTestService(
	txManager *mockTxManager,
	userRepo *mockUserRepo,
	orgRepo *mockOrgRepo,
	roleRepo *mockRoleRepo,
	membershipRepo *mockMembershipRepo,
) *organization.CreateOrganizationService {
	return organization.NewCreateOrganizationService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)
}

func makeOwnerRole() *permission.Role {
	return &permission.Role{
		ID: 1, Name: "owner", OrgName: "__SYSTEM__", IsSystem: true,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func makeTestUser(userID string) *user.User {
	return &user.User{
		ID: userID, ExternalID: "ext-" + userID, Name: "Test User",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

// --- Tests ---

func TestCreateOrganizationService_Execute_Success(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	txManager := &mockTxManager{}
	userRepo := &mockUserRepo{}
	orgRepo := &mockOrgRepo{}
	roleRepo := &mockRoleRepo{}
	membershipRepo := &mockMembershipRepo{}

	userRepo.On("GetByID", ctx, userID).Return(makeTestUser(userID), nil)
	membershipRepo.On("CountByUser", ctx, userID).Return(int64(0), nil)
	orgRepo.On("ExistsByName", ctx, "testorg").Return(false, nil)
	roleRepo.On("GetRoleByNameAndOrg", ctx, "owner", "__SYSTEM__").Return(makeOwnerRole(), nil)
	roleRepo.On("GetRoleByNameAndOrg", ctx, "admin", "__SYSTEM__").
		Return(&permission.Role{ID: 2, Name: "admin", OrgName: "__SYSTEM__", IsSystem: true}, nil)
	roleRepo.On("GetRoleByNameAndOrg", ctx, "editor", "__SYSTEM__").
		Return(&permission.Role{ID: 3, Name: "editor", OrgName: "__SYSTEM__", IsSystem: true}, nil)
	roleRepo.On("GetRoleByNameAndOrg", ctx, "viewer", "__SYSTEM__").
		Return(&permission.Role{ID: 4, Name: "viewer", OrgName: "__SYSTEM__", IsSystem: true}, nil)
	// Transaction succeeds — actual repo calls inside the tx are covered by auto-test
	txManager.On("WithTx", ctx).Return(nil)

	svc := newTestService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)

	input := &organization.CreateOrganizationInput{
		DisplayName:      "Test Organization",
		OrganizationName: "testorg",
		OwnerUserID:      userID,
	}

	_, err := svc.Execute(ctx, input)

	// The outer flow (validation, idempotency check, owner-role lookup) succeeded;
	// the transaction itself was mocked as a no-op.
	assert.NoError(t, err)
	txManager.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestCreateOrganizationService_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()

	txManager := &mockTxManager{}
	userRepo := &mockUserRepo{}
	orgRepo := &mockOrgRepo{}
	roleRepo := &mockRoleRepo{}
	membershipRepo := &mockMembershipRepo{}

	userRepo.On("GetByID", ctx, "non-existent-user").Return(nil, nil)

	svc := newTestService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)

	input := &organization.CreateOrganizationInput{
		DisplayName:      "Test Organization",
		OrganizationName: "testorg",
		OwnerUserID:      "non-existent-user",
	}

	output, err := svc.Execute(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, output)

	bizErr, ok := err.(*bizerrors.BusinessError)
	require.True(t, ok, "Error should be a BusinessError")
	assert.Equal(t, bizerrors.UserNotFound.GetCode(), bizErr.Info().GetCode())
}

func TestCreateOrganizationService_Execute_OrganizationAlreadyExists(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	txManager := &mockTxManager{}
	userRepo := &mockUserRepo{}
	orgRepo := &mockOrgRepo{}
	roleRepo := &mockRoleRepo{}
	membershipRepo := &mockMembershipRepo{}

	userRepo.On("GetByID", ctx, userID).Return(makeTestUser(userID), nil)
	membershipRepo.On("CountByUser", ctx, userID).Return(int64(0), nil)
	orgRepo.On("ExistsByName", ctx, "testorg").Return(true, nil)

	svc := newTestService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)

	input := &organization.CreateOrganizationInput{
		DisplayName:      "Test Organization",
		OrganizationName: "testorg",
		OwnerUserID:      userID,
	}

	output, err := svc.Execute(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, output)

	bizErr, ok := err.(*bizerrors.BusinessError)
	require.True(t, ok, "Error should be a BusinessError")
	assert.Equal(t, bizerrors.OrganizationAlreadyExists.GetCode(), bizErr.Info().GetCode())
}

func TestCreateOrganizationService_Execute_UserAlreadyHasOrganization(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	txManager := &mockTxManager{}
	userRepo := &mockUserRepo{}
	orgRepo := &mockOrgRepo{}
	roleRepo := &mockRoleRepo{}
	membershipRepo := &mockMembershipRepo{}

	existingOrg := &domainOrg.Organization{
		Name:        "firstorg",
		DisplayName: "First Organization",
		OwnerID:     userID,
		Status:      domainOrg.OrgStatusActive,
	}

	userRepo.On("GetByID", ctx, userID).Return(makeTestUser(userID), nil)
	membershipRepo.On("CountByUser", ctx, userID).Return(int64(1), nil)
	orgRepo.On("ListByUser", ctx, userID).Return([]*domainOrg.Organization{existingOrg}, nil)

	svc := newTestService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)

	input := &organization.CreateOrganizationInput{
		DisplayName:      "Second Organization",
		OrganizationName: "secondorg",
		OwnerUserID:      userID,
	}

	output, err := svc.Execute(ctx, input)

	assert.NoError(t, err)
	require.NotNil(t, output)
	assert.True(t, output.AlreadyExisted)
	assert.Equal(t, "firstorg", output.OrganizationID)
	assert.Equal(t, "firstorg", output.OrganizationName)
}

func TestCreateOrganizationService_Execute_RoleNotFound(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	txManager := &mockTxManager{}
	userRepo := &mockUserRepo{}
	orgRepo := &mockOrgRepo{}
	roleRepo := &mockRoleRepo{}
	membershipRepo := &mockMembershipRepo{}

	userRepo.On("GetByID", ctx, userID).Return(makeTestUser(userID), nil)
	membershipRepo.On("CountByUser", ctx, userID).Return(int64(0), nil)
	orgRepo.On("ExistsByName", ctx, "testorg").Return(false, nil)
	// Simulate owner role missing from DB — service will attempt recreation but DB is down
	roleRepo.On("GetRoleByNameAndOrg", ctx, "owner", "__SYSTEM__").Return(nil, nil)
	roleRepo.On("CreateRole", ctx, mock.AnythingOfType("*permission.Role")).Return(fmt.Errorf("db error"))

	svc := newTestService(txManager, userRepo, orgRepo, roleRepo, membershipRepo)

	input := &organization.CreateOrganizationInput{
		DisplayName:      "Test Organization",
		OrganizationName: "testorg",
		OwnerUserID:      userID,
	}

	output, err := svc.Execute(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, output)
}
