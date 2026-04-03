package modeldesign_test

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockModelGroupRepository is a mock implementation of ModelGroupRepository.
type MockModelGroupRepository struct {
	mock.Mock
}

func (m *MockModelGroupRepository) Create(ctx context.Context, group *modeldesign.ModelGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockModelGroupRepository) FindByID(ctx context.Context, id string) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepository) FindByName(
	ctx context.Context, orgName, projectSlug, name string,
) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepository) ListByProject(
	ctx context.Context, orgName, projectSlug string,
) ([]*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepository) Update(ctx context.Context, group *modeldesign.ModelGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockModelGroupRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModelGroupRepository) UpdateModelsGroup(
	ctx context.Context, groupID string, newGroupID *string,
) error {
	args := m.Called(ctx, groupID, newGroupID)
	return args.Error(0)
}

func (m *MockModelGroupRepository) GetTailDisplayOrder(
	ctx context.Context, orgName, projectSlug string,
) (string, error) {
	args := m.Called(ctx, orgName, projectSlug)
	return args.String(0), args.Error(1)
}

// Compile-time check that MockModelGroupRepository implements the interface.
var _ modeldesign.ModelGroupRepository = (*MockModelGroupRepository)(nil)

func TestModelGroupRepository_InterfaceCompliance(t *testing.T) {
	repo := &MockModelGroupRepository{}
	ctx := context.Background()

	group := &modeldesign.ModelGroup{
		ID:           "group-1",
		ProjectScope: project.ProjectScope{OrgName: "org", ProjectSlug: "proj"},
		Name:         "payment",
		DisplayOrder: "a0",
	}

	t.Run("Create", func(t *testing.T) {
		repo.On("Create", ctx, group).Return(nil).Once()
		err := repo.Create(ctx, group)
		assert.NoError(t, err)
	})

	t.Run("FindByID returns group", func(t *testing.T) {
		repo.On("FindByID", ctx, "group-1").Return(group, nil).Once()
		got, err := repo.FindByID(ctx, "group-1")
		assert.NoError(t, err)
		assert.Equal(t, "payment", got.Name)
	})

	t.Run("FindByID returns nil for missing", func(t *testing.T) {
		repo.On("FindByID", ctx, "missing").Return(nil, nil).Once()
		got, err := repo.FindByID(ctx, "missing")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("ListByProject returns groups", func(t *testing.T) {
		groups := []*modeldesign.ModelGroup{group}
		repo.On("ListByProject", ctx, "org", "proj").Return(groups, nil).Once()
		got, err := repo.ListByProject(ctx, "org", "proj")
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("Update", func(t *testing.T) {
		repo.On("Update", ctx, group).Return(nil).Once()
		err := repo.Update(ctx, group)
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		repo.On("Delete", ctx, "group-1").Return(nil).Once()
		err := repo.Delete(ctx, "group-1")
		assert.NoError(t, err)
	})

	t.Run("UpdateModelsGroup", func(t *testing.T) {
		repo.On("UpdateModelsGroup", ctx, "group-1", (*string)(nil)).Return(nil).Once()
		err := repo.UpdateModelsGroup(ctx, "group-1", nil)
		assert.NoError(t, err)
	})

	repo.AssertExpectations(t)
}
