package modeldatabase

import (
	"context"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"

	domaincluster "modelcraft/internal/domain/cluster"
	domaindb "modelcraft/internal/domain/modeldatabase"
)

// ModelDatabaseAppService 管理项目数据库注册的应用服务
type ModelDatabaseAppService struct {
	dbRepo      domaindb.ModelDatabaseRepository
	clusterRepo domaincluster.DatabaseClusterRepository
	connManager *repository.ClusterConnectionManager
}

// NewModelDatabaseAppService creates a new ModelDatabaseAppService.
func NewModelDatabaseAppService(
	dbRepo domaindb.ModelDatabaseRepository,
	clusterRepo domaincluster.DatabaseClusterRepository,
	connManager *repository.ClusterConnectionManager,
) *ModelDatabaseAppService {
	return &ModelDatabaseAppService{dbRepo: dbRepo, clusterRepo: clusterRepo, connManager: connManager}
}

// ListRegistered lists all model databases registered under the project from context.
func (s *ModelDatabaseAppService) ListRegistered(ctx context.Context) ([]*domaindb.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.dbRepo.List(ctx, orgName, projectSlug)
}

// ListRaw lists all non-system databases from the cluster and marks which are registered.
func (s *ModelDatabaseAppService) ListRaw(ctx context.Context) ([]*RawDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	conn, err := s.connManager.GetConnection(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.DatabaseConnectionFailed, "cluster connection not available")
	}
	rawNames, err := listMySQLDatabases(ctx, conn)
	if err != nil {
		return nil, err
	}
	registered, err := s.dbRepo.List(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	registeredSet := make(map[string]bool, len(registered))
	for _, r := range registered {
		registeredSet[r.Name] = true
	}
	result := make([]*RawDatabase, 0, len(rawNames))
	for _, name := range rawNames {
		result = append(result, &RawDatabase{Name: name, IsRegistered: registeredSet[name]})
	}
	return result, nil
}

// Register registers a database under the project cluster.
func (s *ModelDatabaseAppService) Register(ctx context.Context, cmd RegisterCommand) (*domaindb.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	cluster, err := s.clusterRepo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	_, dupErr := s.dbRepo.GetByName(ctx, orgName, projectSlug, cmd.Name)
	if dupErr == nil {
		return nil, bizerrors.NewError(bizerrors.Conflict, "database already registered: "+cmd.Name)
	}
	if !shared.IsNotFoundError(dupErr) {
		return nil, dupErr
	}
	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, err
	}
	db := &domaindb.ModelDatabase{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ClusterID:   cluster.ID,
		Name:        cmd.Name,
		Title:       cmd.Title,
		Description: cmd.Description,
		Mode:        cmd.Mode,
	}
	if err := s.dbRepo.Create(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

// Update updates mutable fields (title, description, mode) of a registered database.
func (s *ModelDatabaseAppService) Update(ctx context.Context, cmd UpdateCommand) (*domaindb.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	db, err := s.dbRepo.GetByID(ctx, orgName, projectSlug, cmd.ID)
	if err != nil {
		return nil, err
	}
	if cmd.Title != nil {
		db.Title = *cmd.Title
	}
	if cmd.Description != nil {
		db.Description = *cmd.Description
	}
	if cmd.Mode != nil {
		db.Mode = *cmd.Mode
	}
	if err := s.dbRepo.Update(ctx, orgName, projectSlug, db); err != nil {
		return nil, err
	}
	return db, nil
}

// Unregister removes a model database registration (soft delete).
func (s *ModelDatabaseAppService) Unregister(ctx context.Context, id string) error {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.dbRepo.Delete(ctx, orgName, projectSlug, id)
}

// RawDatabase represents a MySQL database as seen from the cluster.
type RawDatabase struct {
	Name         string
	IsRegistered bool
}

// RegisterCommand is the input for registering a database.
type RegisterCommand struct {
	Name        string
	Title       string
	Description string
	Mode        domaindb.DatabaseMode
}

// UpdateCommand is the input for updating a registered database.
type UpdateCommand struct {
	ID          string
	Title       *string
	Description *string
	Mode        *domaindb.DatabaseMode
}
