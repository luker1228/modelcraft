package repository_test

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockModelGroupRepo is a mock for the ModelGroupRepository interface used in unit tests.
// Full MySQL integration is tested via integration tests.
type MockModelGroupRepo struct {
	mock.Mock
}

func (m *MockModelGroupRepo) Create(ctx context.Context, group *modeldesign.ModelGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockModelGroupRepo) FindByID(ctx context.Context, id string) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepo) FindByName(
	ctx context.Context, orgName, projectSlug, name string,
) (*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepo) ListByProject(
	ctx context.Context, orgName, projectSlug string,
) ([]*modeldesign.ModelGroup, error) {
	args := m.Called(ctx, orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.ModelGroup), args.Error(1)
}

func (m *MockModelGroupRepo) Update(ctx context.Context, group *modeldesign.ModelGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockModelGroupRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModelGroupRepo) UpdateModelsGroup(
	ctx context.Context, groupID string, newGroupID *string,
) error {
	args := m.Called(ctx, groupID, newGroupID)
	return args.Error(0)
}

func (m *MockModelGroupRepo) GetTailDisplayOrder(
	ctx context.Context, orgName, projectSlug string,
) (string, error) {
	args := m.Called(ctx, orgName, projectSlug)
	return args.String(0), args.Error(1)
}

// Compile-time check.
var _ modeldesign.ModelGroupRepository = (*MockModelGroupRepo)(nil)

func TestModelGroupRepository_MockUsage(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	group := &modeldesign.ModelGroup{
		ID:           "g-1",
		ProjectScope: project.ProjectScope{OrgName: "org", ProjectSlug: "proj"},
		Name:         "payment",
		DisplayOrder: "N",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	t.Run("Create and FindByID", func(t *testing.T) {
		repo := &MockModelGroupRepo{}
		repo.On("Create", ctx, group).Return(nil)
		repo.On("FindByID", ctx, "g-1").Return(group, nil)

		err := repo.Create(ctx, group)
		assert.NoError(t, err)

		got, err := repo.FindByID(ctx, "g-1")
		assert.NoError(t, err)
		assert.Equal(t, "payment", got.Name)
		repo.AssertExpectations(t)
	})

	t.Run("ListByProject returns ordered groups", func(t *testing.T) {
		repo := &MockModelGroupRepo{}
		groups := []*modeldesign.ModelGroup{
			{ID: "g-1", Name: "alpha", DisplayOrder: "A"},
			{ID: "g-2", Name: "beta", DisplayOrder: "N"},
		}
		repo.On("ListByProject", ctx, "org", "proj").Return(groups, nil)

		got, err := repo.ListByProject(ctx, "org", "proj")
		assert.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "alpha", got[0].Name)
		repo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &MockModelGroupRepo{}
		repo.On("Delete", ctx, "g-1").Return(nil)

		err := repo.Delete(ctx, "g-1")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UpdateModelsGroup clears group", func(t *testing.T) {
		repo := &MockModelGroupRepo{}
		repo.On("UpdateModelsGroup", ctx, "g-1", (*string)(nil)).Return(nil)

		err := repo.UpdateModelsGroup(ctx, "g-1", nil)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("GetTailDisplayOrder returns empty for no groups", func(t *testing.T) {
		repo := &MockModelGroupRepo{}
		repo.On("GetTailDisplayOrder", ctx, "org", "proj").Return("", nil)

		tail, err := repo.GetTailDisplayOrder(ctx, "org", "proj")
		assert.NoError(t, err)
		assert.Empty(t, tail)
		repo.AssertExpectations(t)
	})
}
