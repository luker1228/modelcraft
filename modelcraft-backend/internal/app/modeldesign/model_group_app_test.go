package modeldesign_test

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	appmodeldesign "modelcraft/internal/app/modeldesign"

	bizerrors "modelcraft/pkg/bizerrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockGroupRepo struct {
	mock.Mock
}

func (m *mockGroupRepo) Create(ctx context.Context, g *modeldesign.ModelGroup) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepo) FindByID(ctx context.Context, id string) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *mockGroupRepo) FindByName(
	ctx context.Context, orgName, projectSlug, name string,
) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *mockGroupRepo) ListByProject(
	ctx context.Context, orgName, projectSlug string,
) ([]*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.ModelGroup), args.Error(1)
}

func (m *mockGroupRepo) Update(ctx context.Context, g *modeldesign.ModelGroup) error {
	return m.Called(ctx, g).Error(0)
}

func (m *mockGroupRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockGroupRepo) UpdateModelsGroup(ctx context.Context, groupID string, newGroupID *string) error {
	return m.Called(ctx, groupID, newGroupID).Error(0)
}

func (m *mockGroupRepo) GetTailDisplayOrder(ctx context.Context, orgName, projectSlug string) (string, error) {
	args := m.Called(ctx, orgName, projectSlug)
	return args.String(0), args.Error(1)
}

type mockModelRepo struct {
	mock.Mock
}

func (m *mockModelRepo) GetByID(
	ctx context.Context, id string, opts ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.DataModel), args.Error(1)
}

func (m *mockModelRepo) Update(ctx context.Context, model *modeldesign.DataModel) error {
	return m.Called(ctx, model).Error(0)
}

// Stub remaining methods to satisfy interface.
func (m *mockModelRepo) Save(ctx context.Context, orgName string, model *modeldesign.DataModel) error {
	return nil
}

func (m *mockModelRepo) UpdateWithVersion(ctx context.Context, model *modeldesign.DataModel, v int64) (int64, error) {
	return 0, nil
}
func (m *mockModelRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockModelRepo) GetByName(
	ctx context.Context, orgName, db, name, proj string, opts ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	return nil, nil
}

func (m *mockModelRepo) FindByDeploymentStatus(
	ctx context.Context, statuses ...modeldesign.DeploymentStatus,
) ([]modeldesign.DataModel, error) {
	return nil, nil
}

func (m *mockModelRepo) Query(ctx context.Context, q modeldesign.ModelQuery) ([]modeldesign.DataModel, int, error) {
	return nil, 0, nil
}

func (m *mockModelRepo) AddFields(ctx context.Context, orgName string, f []*modeldesign.FieldDefinition) error {
	return nil
}

func (m *mockModelRepo) AddRelationField(ctx context.Context, orgName string, f *modeldesign.FieldDefinition) error {
	return nil
}

func (m *mockModelRepo) GetFieldByModelID(
	ctx context.Context, modelID, name string,
) (*modeldesign.FieldDefinition, error) {
	return nil, nil
}

func (m *mockModelRepo) GetFieldsByModelID(
	ctx context.Context, modelID string,
) ([]*modeldesign.FieldDefinition, error) {
	return nil, nil
}

func (m *mockModelRepo) UpdateField(ctx context.Context, f *modeldesign.FieldDefinition) error {
	return nil
}

func (m *mockModelRepo) BulkUpdateFields(ctx context.Context, f []*modeldesign.FieldDefinition) error {
	return nil
}

func (m *mockModelRepo) UpdateFieldsStatus(ctx context.Context, r ...modeldesign.UpdateFieldsStatusRequest) error {
	return nil
}

func (m *mockModelRepo) DeleteFields(ctx context.Context, modelID string, names []string) error {
	return nil
}

func (m *mockModelRepo) BulkDeleteFields(ctx context.Context, r ...modeldesign.DeleteFieldRequest) error {
	return nil
}

func (m *mockModelRepo) GetTailFieldDisplayOrder(ctx context.Context, modelID string) (string, error) {
	return "", nil
}

// --- Helpers ---

func newGroup(id, name, order string) *modeldesign.ModelGroup {
	return &modeldesign.ModelGroup{
		ID:           id,
		ProjectScope: project.ProjectScope{OrgName: "org", ProjectSlug: "proj"},
		Name:         name,
		DisplayOrder: order,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func newService(
	groupRepo modeldesign.ModelGroupRepository,
	modelRepo modeldesign.ModelRepository,
	txManager repository.TxManager,
) *appmodeldesign.ModelGroupAppService {
	return appmodeldesign.NewModelGroupAppService(groupRepo, modelRepo, txManager)
}

// --- CreateGroup Tests ---

func TestCreateGroup_HappyPath(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	repo.On("FindByName", ctx, "org", "proj", "payment").Return(nil, nil)
	repo.On("GetTailDisplayOrder", ctx, "org", "proj").Return("", nil)
	repo.On("Create", ctx, mock.AnythingOfType("*modeldesign.ModelGroup")).Return(nil)

	group, err := svc.CreateGroup(ctx, appmodeldesign.CreateGroupCommand{
		OrgName:     "org",
		ProjectSlug: "proj",
		Name:        "payment",
	})
	assert.NoError(t, err)
	assert.Equal(t, "payment", group.Name)
	repo.AssertExpectations(t)
}

func TestCreateGroup_InvalidName(t *testing.T) {
	ctx := context.Background()
	svc := newService(&mockGroupRepo{}, nil, nil)

	_, err := svc.CreateGroup(ctx, appmodeldesign.CreateGroupCommand{
		OrgName:     "org",
		ProjectSlug: "proj",
		Name:        "2bad",
	})
	assert.Error(t, err)
	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.ParamInvalid.GetCode(), bizErr.Info().GetCode())
}

func TestCreateGroup_DuplicateName(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	existing := newGroup("g-1", "payment", "N")
	repo.On("FindByName", ctx, "org", "proj", "payment").Return(existing, nil)

	_, err := svc.CreateGroup(ctx, appmodeldesign.CreateGroupCommand{
		OrgName:     "org",
		ProjectSlug: "proj",
		Name:        "payment",
	})
	assert.Error(t, err)
	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.GroupAlreadyExists.GetCode(), bizErr.Info().GetCode())
	repo.AssertExpectations(t)
}

// --- RenameGroup Tests ---

func TestRenameGroup_HappyPath(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	group := newGroup("g-1", "payment", "N")
	repo.On("FindByID", ctx, "g-1").Return(group, nil)
	repo.On("FindByName", ctx, "org", "proj", "payments").Return(nil, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*modeldesign.ModelGroup")).Return(nil)

	updated, err := svc.RenameGroup(ctx, appmodeldesign.RenameGroupCommand{
		OrgName:     "org",
		ProjectSlug: "proj",
		GroupID:     "g-1",
		NewName:     "payments",
	})
	assert.NoError(t, err)
	assert.Equal(t, "payments", updated.Name)
	repo.AssertExpectations(t)
}

func TestRenameGroup_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	repo.On("FindByID", ctx, "missing").Return(nil, nil)

	_, err := svc.RenameGroup(ctx, appmodeldesign.RenameGroupCommand{
		GroupID: "missing",
		NewName: "anything",
	})
	assert.Error(t, err)
	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.GroupNotFound.GetCode(), bizErr.Info().GetCode())
}

func TestRenameGroup_ConflictName(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	group := newGroup("g-1", "payment", "N")
	conflict := newGroup("g-2", "payments", "Z")
	repo.On("FindByID", ctx, "g-1").Return(group, nil)
	repo.On("FindByName", ctx, "org", "proj", "payments").Return(conflict, nil)

	_, err := svc.RenameGroup(ctx, appmodeldesign.RenameGroupCommand{
		OrgName:     "org",
		ProjectSlug: "proj",
		GroupID:     "g-1",
		NewName:     "payments",
	})
	assert.Error(t, err)
	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.GroupAlreadyExists.GetCode(), bizErr.Info().GetCode())
}

// --- DeleteGroup Tests ---

func TestDeleteGroup_WithModels(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	repo.On("FindByID", ctx, "g-1").Return(newGroup("g-1", "payment", "N"), nil)
	repo.On("UpdateModelsGroup", ctx, "g-1", (*string)(nil)).Return(nil)
	repo.On("Delete", ctx, "g-1").Return(nil)

	err := svc.DeleteGroup(ctx, "g-1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteGroup_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	repo.On("FindByID", ctx, "missing").Return(nil, nil)

	err := svc.DeleteGroup(ctx, "missing")
	assert.Error(t, err)
	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.GroupNotFound.GetCode(), bizErr.Info().GetCode())
}

// --- ReorderGroup Tests ---

func TestReorderGroup_ToHead(t *testing.T) {
	ctx := context.Background()
	repo := &mockGroupRepo{}
	svc := newService(repo, nil, nil)

	group := newGroup("g-1", "payment", "N")
	firstGroup := newGroup("g-2", "alpha", "!")
	groups := []*modeldesign.ModelGroup{firstGroup, group}

	repo.On("FindByID", ctx, "g-1").Return(group, nil)
	repo.On("ListByProject", ctx, "org", "proj").Return(groups, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*modeldesign.ModelGroup")).Return(nil)

	err := svc.ReorderGroup(ctx, appmodeldesign.ReorderGroupCommand{
		OrgName:      "org",
		ProjectSlug:  "proj",
		GroupID:      "g-1",
		AfterGroupID: nil, // move to head
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// --- MoveModelToGroup Tests ---

func TestMoveModelToGroup_HappyPath(t *testing.T) {
	ctx := context.Background()
	groupRepo := &mockGroupRepo{}
	modelRepo := &mockModelRepo{}
	svc := newService(groupRepo, modelRepo, nil)

	groupID := "g-1"
	group := newGroup(groupID, "payment", "N")
	model := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: "m-1",
		},
	}
	model.ModelLocator = modeldesign.ModelLocator{ProjectScope: project.ProjectScope{ProjectSlug: "proj"}}

	groupRepo.On("FindByID", ctx, groupID).Return(group, nil)
	modelRepo.On("GetByID", ctx, "m-1").Return(model, nil)
	modelRepo.On("Update", ctx, mock.AnythingOfType("*modeldesign.DataModel")).Return(nil)

	err := svc.MoveModelToGroup(ctx, appmodeldesign.MoveModelToGroupCommand{
		ModelID: "m-1",
		GroupID: &groupID,
	})
	assert.NoError(t, err)
	repo := modelRepo
	repo.AssertExpectations(t)
}

func TestMoveModelToGroup_Ungrouped(t *testing.T) {
	ctx := context.Background()
	groupRepo := &mockGroupRepo{}
	modelRepo := &mockModelRepo{}
	svc := newService(groupRepo, modelRepo, nil)

	model := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: "m-1",
		},
	}
	model.ModelLocator = modeldesign.ModelLocator{ProjectScope: project.ProjectScope{ProjectSlug: "proj"}}

	modelRepo.On("GetByID", ctx, "m-1").Return(model, nil)
	modelRepo.On("Update", ctx, mock.AnythingOfType("*modeldesign.DataModel")).Return(nil)

	err := svc.MoveModelToGroup(ctx, appmodeldesign.MoveModelToGroupCommand{
		ModelID: "m-1",
		GroupID: nil, // move to ungrouped
	})
	assert.NoError(t, err)
	modelRepo.AssertExpectations(t)
}

// --- ListGroups Tests ---

func TestListGroups_WithUngrouped(t *testing.T) {
	ctx := context.Background()
	groupRepo := &mockGroupRepo{}
	modelRepo := &mockModelRepo{}
	svc := newService(groupRepo, modelRepo, nil)

	groups := []*modeldesign.ModelGroup{
		newGroup("g-1", "payment", "N"),
	}
	groupRepo.On("ListByProject", ctx, "org", "proj").Return(groups, nil)

	result, err := svc.ListGroups(ctx, "org", "proj")
	assert.NoError(t, err)
	// Should include real groups + virtual ungrouped at end
	assert.Len(t, result, 2)
	assert.Equal(t, "payment", result[0].Name)
	assert.Equal(t, modeldesign.UngroupedGroupID, result[1].ID)
	assert.True(t, result[1].IsVirtual())
	groupRepo.AssertExpectations(t)
}

func TestListGroups_Empty(t *testing.T) {
	ctx := context.Background()
	groupRepo := &mockGroupRepo{}
	svc := newService(groupRepo, nil, nil)

	groupRepo.On("ListByProject", ctx, "org", "proj").Return([]*modeldesign.ModelGroup{}, nil)

	result, err := svc.ListGroups(ctx, "org", "proj")
	assert.NoError(t, err)
	// Only virtual ungrouped
	assert.Len(t, result, 1)
	assert.True(t, result[0].IsVirtual())
}
