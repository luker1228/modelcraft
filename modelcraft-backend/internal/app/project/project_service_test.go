package project

import (
	"context"
	"modelcraft/internal/domain/project"
	"modelcraft/pkg/bizerrors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository is a mock implementation of project.ProjectRepository.
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, proj *project.Project) error {
	args := m.Called(ctx, proj)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByNameAndOrg(
	ctx context.Context,
	name string,
	orgName string,
) (*project.Project, error) {
	args := m.Called(ctx, name, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

func (m *MockProjectRepository) List(
	ctx context.Context,
	status *project.ProjectStatus,
) ([]*project.Project, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.Project), args.Error(1)
}

func (m *MockProjectRepository) ListByOrg(
	ctx context.Context,
	orgName string,
	status *project.ProjectStatus,
) ([]*project.Project, error) {
	args := m.Called(ctx, orgName, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*project.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(ctx context.Context, proj *project.Project) error {
	args := m.Called(ctx, proj)
	return args.Error(0)
}

func (m *MockProjectRepository) Archive(ctx context.Context, name, orgName string) error {
	args := m.Called(ctx, name, orgName)
	return args.Error(0)
}

func (m *MockProjectRepository) ExistsByName(ctx context.Context, name, orgName string) (bool, error) {
	args := m.Called(ctx, name, orgName)
	return args.Bool(0), args.Error(1)
}

func (m *MockProjectRepository) GetByClusterID(
	ctx context.Context,
	orgName string,
	clusterID string,
) (*project.Project, error) {
	args := m.Called(ctx, orgName, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*project.Project), args.Error(1)
}

// newTestService creates a service with nil cluster deps (sufficient for non-transactional tests).
func newTestService(mockRepo *MockProjectRepository) *ProjectAppService {
	return NewProjectAppService(mockRepo, nil, nil, nil)
}

func TestProjectAppService_CreateProject_AlreadyExists(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockProjectRepository)
	service := newTestService(mockRepo)

	mockRepo.On("ExistsByName", ctx, "existing_project", "built-in").Return(true, nil)

	cmd := CreateProjectCommand{
		OrgName:            "built-in",
		Slug:               "existing_project",
		Title:              "Test",
		Description:        "Description",
		SkipConnectionTest: true,
		ClusterInput: CreateClusterForProjectInput{
			Title:    "Cluster 1",
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "pass",
		},
	}
	proj, err := service.CreateProject(ctx, cmd)

	assert.Error(t, err)
	assert.Nil(t, proj)

	var bizErr *bizerrors.BusinessError
	assert.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ProjectAlreadyExists.GetCode(), bizErr.Info().GetCode())
	mockRepo.AssertExpectations(t)
}

func TestProjectAppService_GetProject(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		expectedProj, _ := project.NewProject("built-in", "test_project", "Test", "Description", "")
		mockRepo.On("GetByNameAndOrg", ctx, "test_project", "built-in").Return(expectedProj, nil)

		cmd := GetProjectCommand{
			OrgName: "built-in",
			Slug:    "test_project",
		}
		proj, err := service.GetProjectByNameAndOrg(ctx, cmd)

		assert.NoError(t, err)
		assert.NotNil(t, proj)
		assert.Equal(t, "test_project", proj.Slug)
		mockRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		mockRepo.On("GetByNameAndOrg", ctx, "nonexistent", "built-in").Return(nil, nil)

		cmd := GetProjectCommand{
			OrgName: "built-in",
			Slug:    "nonexistent",
		}
		proj, err := service.GetProjectByNameAndOrg(ctx, cmd)

		assert.Error(t, err)
		assert.Nil(t, proj)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.NotFound, bizErr.Info())
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectAppService_ListProjects(t *testing.T) {
	ctx := context.Background()

	t.Run("list all", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		proj1, _ := project.NewProject("built-in", "project1", "Project 1", "Desc 1", "")
		proj2, _ := project.NewProject("built-in", "project2", "Project 2", "Desc 2", "")
		expectedProjects := []*project.Project{proj1, proj2}

		mockRepo.On("List", ctx, (*project.ProjectStatus)(nil)).Return(expectedProjects, nil)

		projects, err := service.ListProjects(ctx, nil)

		assert.NoError(t, err)
		assert.Len(t, projects, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list by status", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		proj1, _ := project.NewProject("built-in", "project1", "Project 1", "Desc 1", "")
		expectedProjects := []*project.Project{proj1}

		activeStatus := project.ProjectStatusActive
		mockRepo.On("List", ctx, &activeStatus).Return(expectedProjects, nil)

		projects, err := service.ListProjects(ctx, &activeStatus)

		assert.NoError(t, err)
		assert.Len(t, projects, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectAppService_UpdateProjectMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		existingProj, _ := project.NewProject("built-in", "test_project", "Old Title", "Old Description", "")
		mockRepo.On("GetByNameAndOrg", ctx, "test_project", "built-in").Return(existingProj, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*project.Project")).Return(nil)

		cmd := UpdateProjectCommand{
			OrgName:     "built-in",
			Slug:        "test_project",
			Title:       "New Title",
			Description: "New Description",
			LoginURL:    "",
		}
		updatedProj, err := service.UpdateProjectMetadata(ctx, cmd)

		assert.NoError(t, err)
		assert.NotNil(t, updatedProj)
		assert.Equal(t, "New Title", updatedProj.Title)
		assert.Equal(t, "New Description", updatedProj.Description)
		mockRepo.AssertExpectations(t)
	})

	t.Run("project not found", func(t *testing.T) {
		mockRepo := new(MockProjectRepository)
		service := newTestService(mockRepo)

		mockRepo.On("GetByNameAndOrg", ctx, "nonexistent", "built-in").Return(nil, nil)

		cmd := UpdateProjectCommand{
			OrgName:     "built-in",
			Slug:        "nonexistent",
			Title:       "Title",
			Description: "Desc",
			LoginURL:    "",
		}
		proj, err := service.UpdateProjectMetadata(ctx, cmd)

		assert.Error(t, err)
		assert.Nil(t, proj)
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectAppService_DeleteProject_ProjectNotFound(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockProjectRepository)
	service := newTestService(mockRepo)

	mockRepo.On("GetByNameAndOrg", ctx, "nonexistent", "built-in").Return(nil, nil)

	cmd := DeleteProjectCommand{
		OrgName: "built-in",
		Slug:    "nonexistent",
	}
	err := service.DeleteProject(ctx, cmd)

	assert.Error(t, err)
	var bizErr *bizerrors.BusinessError
	assert.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.NotFound, bizErr.Info())
	mockRepo.AssertExpectations(t)
}
