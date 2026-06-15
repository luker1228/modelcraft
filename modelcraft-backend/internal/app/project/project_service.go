package project

import (
	"context"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"

	clusterApp "modelcraft/internal/app/cluster"
	domainCluster "modelcraft/internal/domain/cluster"
)

// PrivateDBProvisioner is deprecated and kept only for compatibility.
// Implemented by infrastructure PrivateDBManager.
type PrivateDBProvisioner interface {
	Provision(ctx context.Context, orgName, projectSlug string) error
}

// PrivateModelImporter is deprecated and kept only for compatibility.
// Implemented by modeldesign.ReverseEngineerAppService.
type PrivateModelImporter interface {
	ImportPrivateModel(ctx context.Context, orgName, projectSlug, databaseName, tableName string) error
}

// ProjectAppService orchestrates project lifecycle use cases.
type ProjectAppService struct {
	projectRepo    project.ProjectRepository
	clusterRepo    domainCluster.DatabaseClusterRepository
	clusterService *clusterApp.DatabaseClusterAppService
	txManager      repository.TxManager
	privateDB      PrivateDBProvisioner // optional: nil skips provisioning
	modelImporter  PrivateModelImporter // optional: nil skips model import
}

// NewProjectAppService creates a new ProjectAppService with all required dependencies.
func NewProjectAppService(
	projectRepo project.ProjectRepository,
	clusterRepo domainCluster.DatabaseClusterRepository,
	clusterService *clusterApp.DatabaseClusterAppService,
	txManager repository.TxManager,
) *ProjectAppService {
	return &ProjectAppService{
		projectRepo:    projectRepo,
		clusterRepo:    clusterRepo,
		clusterService: clusterService,
		txManager:      txManager,
	}
}

// WithPrivateDBProvisioner sets the PrivateDBProvisioner used to provision
// mc_private_{projectSlug} on the project's cluster after creation.
func (s *ProjectAppService) WithPrivateDBProvisioner(p PrivateDBProvisioner) *ProjectAppService {
	s.privateDB = p
	return s
}

// WithPrivateModelImporter sets the PrivateModelImporter used to import
// users and accounts tables as models after private DB is provisioned.
func (s *ProjectAppService) WithPrivateModelImporter(i PrivateModelImporter) *ProjectAppService {
	s.modelImporter = i
	return s
}

// CreateProject atomically creates a project and its associated cluster.
// Connection is tested before the transaction unless SkipConnectionTest is true.
func (s *ProjectAppService) CreateProject(
	ctx context.Context,
	cmd CreateProjectCommand,
) (*project.Project, error) {
	logger := logfacade.GetLogger(ctx)

	// Check project name uniqueness
	exists, err := s.projectRepo.ExistsByName(ctx, cmd.Slug, cmd.OrgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to check project existence")
	}
	if exists {
		return nil, bizerrors.NewError(
			bizerrors.ProjectAlreadyExists,
			"project '%s' already exists in organization '%s'",
			cmd.Slug,
			cmd.OrgName,
		)
	}

	// Test connection before starting the transaction (avoids holding tx open during network I/O)
	if !cmd.SkipConnectionTest {
		ct := cmd.ClusterInput.ConnectionTimeout
		testCmd := &clusterApp.TestConnectionCommand{
			Host:              cmd.ClusterInput.Host,
			Port:              cmd.ClusterInput.Port,
			Username:          cmd.ClusterInput.Username,
			Password:          cmd.ClusterInput.Password,
			ConnectionTimeout: &ct,
		}
		if err := s.clusterService.TestConnection(ctx, cmd.Slug, nil, testCmd); err != nil {
			return nil, err
		}
	}

	// Create project entity
	proj, err := project.NewProject(cmd.OrgName, cmd.Slug, cmd.Title, cmd.Description)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to create project entity")
	}

	// Generate cluster ID upfront so we can reference it in the transaction
	clusterID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to generate cluster ID")
	}

	// Atomically persist project and cluster
	txErr := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		txProjectRepo := repository.NewSqlProjectRepository(q)
		txClusterRepo := repository.NewSqlDatabaseClusterRepository(q)

		// Save project
		if err := txProjectRepo.Create(ctx, proj); err != nil {
			return bizerrors.Wrapf(err, "failed to save project")
		}

		// Build and save cluster entity
		clusterEntity, buildErr := domainCluster.NewDatabaseCluster(
			cmd.OrgName,
			cmd.Slug,
			cmd.ClusterInput.Title,
			cmd.ClusterInput.Host,
			cmd.ClusterInput.Port,
			cmd.ClusterInput.Username,
			cmd.ClusterInput.Password,
		)
		if buildErr != nil {
			return buildErr
		}
		clusterEntity.ID = clusterID
		if cmd.ClusterInput.Description != "" {
			clusterEntity.Description = cmd.ClusterInput.Description
		}
		if cmd.ClusterInput.ConnectionTimeout != 0 {
			clusterEntity.ConnectionTimeout = cmd.ClusterInput.ConnectionTimeout
		}
		clusterEntity.SetDefaults()

		if err := clusterEntity.Validate(); err != nil {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, err.Error())
		}

		if err := txClusterRepo.Create(ctx, clusterEntity); err != nil {
			return bizerrors.Wrapf(err, "failed to save cluster")
		}

		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	logger.Infof(ctx, "project and cluster created successfully: project=%s cluster=%s", proj.Slug, clusterID)

	return proj, nil
}

// GetProjectByNameAndOrg retrieves a project by name within an organization.
func (s *ProjectAppService) GetProjectByNameAndOrg(
	ctx context.Context,
	cmd GetProjectCommand,
) (*project.Project, error) {
	proj, err := s.projectRepo.GetByNameAndOrg(ctx, cmd.Slug, cmd.OrgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get project")
	}
	if proj == nil {
		return nil, bizerrors.NewError(
			bizerrors.NotFound,
			"project '%s' not found in organization '%s'",
			cmd.Slug,
			cmd.OrgName,
		)
	}
	return proj, nil
}

// ListProjects retrieves all projects, optionally filtered by status.
func (s *ProjectAppService) ListProjects(
	ctx context.Context,
	status *project.ProjectStatus,
) ([]*project.Project, error) {
	projects, err := s.projectRepo.List(ctx, status)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list projects")
	}
	return projects, nil
}

// ListProjectsByOrg retrieves projects for a specific organization.
func (s *ProjectAppService) ListProjectsByOrg(
	ctx context.Context,
	cmd ListProjectsCommand,
) ([]*project.Project, error) {
	projects, err := s.projectRepo.ListByOrg(ctx, cmd.OrgName, cmd.Status)
	if err != nil {
		return nil, bizerrors.Wrapf(
			err,
			"failed to list projects for organization '%s'",
			cmd.OrgName,
		)
	}
	return projects, nil
}

// UpdateProjectMetadata updates the mutable metadata fields of a project.
func (s *ProjectAppService) UpdateProjectMetadata(
	ctx context.Context,
	cmd UpdateProjectCommand,
) (*project.Project, error) {
	proj, err := s.projectRepo.GetByNameAndOrg(ctx, cmd.Slug, cmd.OrgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get project")
	}
	if proj == nil {
		return nil, bizerrors.NewError(
			bizerrors.NotFound,
			"project '%s' not found in organization '%s'",
			cmd.Slug,
			cmd.OrgName,
		)
	}

	if err := proj.UpdateMetadata(cmd.Title, cmd.Description); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to update project metadata")
	}

	if err := s.projectRepo.Update(ctx, proj); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to save project")
	}

	return proj, nil
}

// DeleteProject archives a project and cascade-deletes its cluster in a single transaction.
func (s *ProjectAppService) DeleteProject(
	ctx context.Context,
	cmd DeleteProjectCommand,
) error {
	logger := logfacade.GetLogger(ctx)

	// Verify the project exists before starting the transaction
	proj, err := s.projectRepo.GetByNameAndOrg(ctx, cmd.Slug, cmd.OrgName)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to get project")
	}
	if proj == nil {
		return bizerrors.NewError(
			bizerrors.NotFound,
			"project '%s' not found in organization '%s'",
			cmd.Slug,
			cmd.OrgName,
		)
	}

	txErr := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		txProjectRepo := repository.NewSqlProjectRepository(q)
		txClusterRepo := repository.NewSqlDatabaseClusterRepository(q)

		// Delete cluster if it exists (cascade)
		clusterEntity, clusterErr := txClusterRepo.GetByProjectKey(ctx, cmd.OrgName, cmd.Slug)
		if clusterErr == nil && clusterEntity != nil {
			if err := txClusterRepo.Delete(ctx, cmd.OrgName, proj.Slug, clusterEntity.ID); err != nil {
				return bizerrors.Wrapf(err, "failed to delete cluster for project")
			}
		}

		// Archive the project
		if err := txProjectRepo.Archive(ctx, cmd.Slug, cmd.OrgName); err != nil {
			return bizerrors.Wrapf(err, "failed to archive project")
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	logger.Infof(ctx, "project deleted (archived) with cascade cluster deletion: %s", cmd.Slug)
	return nil
}

// EnsureDefaultProject is a no-op. Default project auto-creation is disabled
// because a cluster is now required when creating a project.
func (s *ProjectAppService) EnsureDefaultProject(_ context.Context) error {
	return nil
}
