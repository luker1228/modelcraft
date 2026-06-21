package cluster

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
)

// DatabaseClusterAppService represents the database cluster application service
type DatabaseClusterAppService struct {
	clusterRepo cluster.DatabaseClusterRepository
	connManager *repository.ClusterConnectionManager
}

// NewDatabaseClusterAppService creates a new database cluster application service
func NewDatabaseClusterAppService(
	clusterRepo cluster.DatabaseClusterRepository,
	connManager *repository.ClusterConnectionManager,
) *DatabaseClusterAppService {
	return &DatabaseClusterAppService{
		clusterRepo: clusterRepo,
		connManager: connManager,
	}
}

// getLogger retrieves the logger instance
func (s *DatabaseClusterAppService) getLogger(ctx context.Context) logfacade.Logger {
	return logfacade.GetLogger(ctx)
}

// getOrgNameFromContext extracts orgName from context
// Returns error if orgName is not found in context
func getOrgNameFromContext(ctx context.Context) (string, error) {
	return ctxutils.GetOrgNameFromContext(ctx)
}

// CreateCluster creates a new database cluster
func (s *DatabaseClusterAppService) CreateCluster(ctx context.Context, cmd CreateClusterCommand) (string, error) {
	logger := s.getLogger(ctx)

	// Check one-to-one constraint: project should not already have a cluster
	hasCluster, err := s.clusterRepo.ExistsByProjectKey(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		logger.Errorf(ctx, "failed to check if project has cluster, err=%v", err)
		return "", bizerrors.Wrapf(err, "failed to check if project has cluster")
	}
	if hasCluster {
		logger.Infof(ctx, "project %s/%s already has a cluster", cmd.OrgName, cmd.ProjectSlug)
		return "", bizerrors.NewError(bizerrors.ProjectAlreadyHasCluster, cmd.ProjectSlug)
	}

	connectionTimeout := cmd.ConnectionTimeout
	err = s.TestConnection(ctx, cmd.ProjectSlug, nil, &TestConnectionCommand{
		Host:              cmd.Host,
		Port:              cmd.Port,
		Username:          cmd.Username,
		Password:          cmd.Password,
		ConnectionTimeout: &connectionTimeout,
	})
	if err != nil {
		return "", err
	}

	// Create entity
	entity, err := cluster.NewDatabaseCluster(
		cmd.OrgName,
		cmd.ProjectSlug,
		cmd.Title,
		cmd.Host,
		cmd.Port,
		cmd.Username,
		cmd.Password,
	)
	if err != nil {
		return "", err
	}

	// Set optional fields
	if cmd.Description != "" {
		entity.Description = cmd.Description
	}
	if cmd.ConnectionTimeout != 0 {
		entity.ConnectionTimeout = cmd.ConnectionTimeout
	}

	// Generate UUIDV7 (naturally ordered)
	uuid, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return "", err
	}
	entity.ID = uuid

	// Validate entity
	if err := entity.Validate(); err != nil {
		logger.Errorf(ctx, "database cluster validation failed: %v", err)
		return "", bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, err.Error())
	}

	// Save to database
	if err := s.clusterRepo.Create(ctx, entity); err != nil {
		logger.Errorf(ctx, "failed to create database cluster: %v", err)
		return "", err
	}

	logger.Infof(ctx, "database cluster created successfully: %s", entity.ID)
	return entity.ID, nil
}

// UpdateProjectCluster updates the database cluster connection for a project.
// Since each project has exactly one cluster, the cluster is identified by project name.
func (s *DatabaseClusterAppService) UpdateProjectCluster(
	ctx context.Context,
	cmd UpdateProjectClusterCommand,
) (*cluster.DatabaseCluster, error) {
	logger := s.getLogger(ctx)

	// Get existing cluster by project key (one-to-one relationship)
	entity, err := s.clusterRepo.GetByProjectKey(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		logger.Errorf(ctx, "Failed to get cluster by project key, err=%v", err)
		return nil, bizerrors.Wrapf(err, "failed to get cluster for project")
	}
	if entity == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound, cmd.ProjectSlug)
	}

	// Apply partial updates
	if cmd.Title != nil {
		entity.Title = *cmd.Title
	}
	if cmd.Description != nil {
		entity.Description = *cmd.Description
	}
	if cmd.Host != nil {
		entity.Host = *cmd.Host
	}
	if cmd.Port != nil {
		entity.Port = *cmd.Port
	}
	if cmd.Username != nil {
		entity.Username = *cmd.Username
	}
	if cmd.Password != nil && *cmd.Password != cluster.EncryptedByServerPlaceholder {
		plainPasswd, err := cluster.NewByPlain(*cmd.Password)
		if err != nil {
			return nil, err
		}
		entity.Password = *plainPasswd
	}
	if cmd.ConnectionTimeout != nil {
		entity.ConnectionTimeout = *cmd.ConnectionTimeout
	}

	entity.SetDefaults()

	// Validate updated entity
	if err := entity.Validate(); err != nil {
		logger.Errorf(ctx, "database cluster validation failed: %v", err)
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, err.Error())
	}

	// Test connection unless explicitly skipped
	if !cmd.SkipConnectionTest {
		connInfo := entity.GetConnectionInfo()
		if err := normalizeConnectionTimeout(ctx, &connInfo); err != nil {
			return nil, err
		}
		if err := s.connManager.TestConnection(ctx, connInfo); err != nil {
			logger.Errorf(ctx, "connection test failed: %v", err)
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.DatabaseConnectionFailed, err.Error())
		}
	}

	// Persist
	if err := s.clusterRepo.Update(ctx, cmd.OrgName, cmd.ProjectSlug, entity); err != nil {
		logger.Errorf(ctx, "Failed to update database cluster, err=%v", err)
		return nil, bizerrors.Wrapf(err, "failed to update cluster")
	}

	logger.Infof(ctx, "database cluster updated successfully for project: %s", cmd.ProjectSlug)
	return entity, nil
}

// GetCluster retrieves database cluster details by ID
func (s *DatabaseClusterAppService) GetCluster(
	ctx context.Context,
	projectID, id string,
) (*cluster.DatabaseCluster, error) {
	logger := s.getLogger(ctx)

	// Get orgName from context
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entity, err := s.clusterRepo.GetByID(ctx, orgName, id)
	if err != nil {
		logger.Errorf(ctx, "failed to get database cluster: %v", err)
		return nil, err
	}

	return entity, nil
}

// GetClusterByProject retrieves the database cluster for a project by project slug.
// Each project has at most one cluster (one-to-one relationship).
// Returns ClusterNotFound error if no cluster exists for the project.
func (s *DatabaseClusterAppService) GetClusterByProject(
	ctx context.Context,
	projectSlug string,
) (*cluster.DatabaseCluster, error) {
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entity, err := s.clusterRepo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to get database cluster by project")
	}
	if entity == nil {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ClusterNotFound,
			projectSlug,
		)
	}

	return entity, nil
}

// ListClusters retrieves the list of database clusters
func (s *DatabaseClusterAppService) ListClusters(
	ctx context.Context,
	projectID string,
) ([]*cluster.DatabaseCluster, error) {
	logger := s.getLogger(ctx)

	// Get orgName from context
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	entities, err := s.clusterRepo.List(ctx, orgName, projectID)
	if err != nil {
		logger.Errorf(ctx, "failed to get database cluster list: %v", err)
		return nil, err
	}

	return entities, nil
}

// DeleteCluster deletes a database cluster
func (s *DatabaseClusterAppService) DeleteCluster(ctx context.Context, projectID, id string) error {
	logger := s.getLogger(ctx)

	// Get orgName from context
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return err
	}

	// TODO: Check if any models are using this cluster

	// Delete cluster
	if err := s.clusterRepo.Delete(ctx, orgName, projectID, id); err != nil {
		logger.Errorf(ctx, "failed to delete database cluster: %v", err)
		return err
	}

	// Close connection (ignore error since deletion is already successful)
	_ = s.connManager.RefreshConnection(ctx, orgName, projectID)

	logger.Infof(ctx, "database cluster deleted successfully: %s", id)
	return nil
}

// TestConnection tests database connection
func (s *DatabaseClusterAppService) TestConnection(
	ctx context.Context, projectID string, id *string, cmd *TestConnectionCommand,
) error {
	logger := s.getLogger(ctx)

	var connInfo *cluster.ConnectionInfo

	// If ID is provided, get cluster information from database
	if id != nil {
		info, err := s.buildConnInfoFromID(ctx, *id)
		if err != nil {
			return err
		}
		connInfo = info
	}

	// If connection info is provided, use it to override fields from the database cluster.
	// When cmd.Password is empty (sentinel "<encrypted_by_server>" stripped by caller),
	// the password from the database cluster (loaded above via id) is preserved.
	if cmd != nil {
		info, err := buildConnInfoFromCmd(cmd, connInfo)
		if err != nil {
			return err
		}
		connInfo = info
	}

	// Check if connection info is available
	if connInfo == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid,
			"cluster ID or connection info must be provided")
	}

	if err := normalizeConnectionTimeout(ctx, connInfo); err != nil {
		return err
	}

	// Test connection
	if s.connManager == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "connection manager not available")
	}
	err := s.connManager.TestConnection(ctx, *connInfo)
	if err != nil {
		logger.Errorf(ctx, "connection test failed: %v", err)
		return bizerrors.NewErrorFromContext(ctx, bizerrors.DatabaseConnectionFailed, err.Error())
	}

	logger.Infof(ctx, "database connection test successful")
	return nil
}

func (s *DatabaseClusterAppService) buildConnInfoFromID(
	ctx context.Context, id string,
) (*cluster.ConnectionInfo, error) {
	logger := s.getLogger(ctx)
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := s.clusterRepo.GetByID(ctx, orgName, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound,
				fmt.Sprintf("cluster not found id=%s", id))
		}
		logger.Errorf(ctx, "failed to get database cluster: %v", err)
		return nil, err
	}
	if clusterInfo == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound,
			fmt.Sprintf("cluster not found id=%s", id))
	}

	clusterInfo.SetDefaults()
	info := clusterInfo.GetConnectionInfo()
	return &info, nil
}

// buildConnInfoFromCmd builds a ConnectionInfo from a TestConnectionCommand.
// If cmd.Password is empty, the password is taken from fallback (i.e. the existing cluster's
// connection info loaded from the database). All other fields always come from cmd.
func buildConnInfoFromCmd(cmd *TestConnectionCommand,
	fallback *cluster.ConnectionInfo,
) (*cluster.ConnectionInfo, error) {
	var passwd cluster.Password

	if cmd.Password == "" {
		// Use the database password from the existing cluster.
		if fallback == nil {
			return nil, bizerrors.Errorf("password is required when no existing cluster is available")
		}
		passwd = fallback.Password
	} else {
		passwdobj, err := cluster.NewByPlain(cmd.Password)
		if err != nil {
			return nil, err
		}
		passwd = *passwdobj
	}

	connectionTimeout := 0
	if cmd.ConnectionTimeout != nil {
		connectionTimeout = *cmd.ConnectionTimeout
	}

	return &cluster.ConnectionInfo{
		Username:          cmd.Username,
		Password:          passwd,
		Host:              cmd.Host,
		Port:              cmd.Port,
		ConnectionTimeout: connectionTimeout,
	}, nil
}

func normalizeConnectionTimeout(ctx context.Context, connInfo *cluster.ConnectionInfo) error {
	if connInfo.ConnectionTimeout == 0 {
		connInfo.ConnectionTimeout = 5
	}
	if connInfo.ConnectionTimeout < 5 || connInfo.ConnectionTimeout > 15 {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"connection timeout must be between 5 and 15 seconds",
		)
	}
	return nil
}

// DatabaseInfo represents database information
type DatabaseInfo struct {
	Name string
}

// ListDatabases lists all databases in the cluster with pagination
func (s *DatabaseClusterAppService) ListDatabases(
	ctx context.Context,
	projectID, search string,
	offset, limit int,
) ([]*DatabaseInfo, int, error) {
	logger := s.getLogger(ctx)

	// Get orgName from context
	orgName, err := getOrgNameFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Call connection manager's ListDatabases method with pagination
	databases, totalCount, err := s.connManager.ListDatabases(ctx, orgName, projectID, search, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "failed to list databases: %v", err)
		return nil, 0, bizerrors.NewErrorFromContext(ctx, bizerrors.DatabaseConnectionFailed, err.Error())
	}

	// Convert to application layer type
	result := make([]*DatabaseInfo, len(databases))
	for i, db := range databases {
		result[i] = &DatabaseInfo{Name: db.Name}
	}

	logger.Infof(ctx, "successfully listed databases, total: %d, returned: %d", totalCount, len(result))
	return result, totalCount, nil
}
