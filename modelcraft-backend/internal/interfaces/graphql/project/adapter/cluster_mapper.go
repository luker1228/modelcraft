package adapter

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizutils"
	"time"

	clusterApp "modelcraft/internal/app/cluster"
	domainCluster "modelcraft/internal/domain/cluster"
)

// ClusterMapper maps between domain cluster models and GraphQL types.
type ClusterMapper struct{}

// NewClusterMapper creates a new ClusterMapper.
func NewClusterMapper() *ClusterMapper {
	return &ClusterMapper{}
}

// ToGraphQLCluster converts a domain DatabaseCluster to its GraphQL representation.
func (m *ClusterMapper) ToGraphQLCluster(domain *domainCluster.DatabaseCluster) *generated.DatabaseCluster {
	if domain == nil {
		return nil
	}

	connectionInfo := generated.DatabaseConnectionInfo{
		Host:     domain.Host,
		Port:     domain.Port,
		Username: domain.Username,
		Password: domain.Password.GetViewPasswd(),
	}

	return &generated.DatabaseCluster{
		ID:                domain.ID,
		ProjectSlug:       domain.ProjectSlug,
		Title:             domain.Title,
		Description:       domain.Description,
		ConnectionInfo:    &connectionInfo,
		ConnectionTimeout: int32(domain.ConnectionTimeout), //nolint:gosec // G115: timeout is small
		Status:            m.toGraphQLStatus(domain.Status),
		Version:           int32(domain.Version), //nolint:gosec // G115: version won't exceed int32 in practice
		CreatedAt:         domain.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         domain.UpdatedAt.Format(time.RFC3339),
		DeletedAt:         nil,
	}
}

// ToGraphQLClusters converts a slice of domain DatabaseClusters to GraphQL types.
func (m *ClusterMapper) ToGraphQLClusters(domains []*domainCluster.DatabaseCluster) []*generated.DatabaseCluster {
	if domains == nil {
		return nil
	}

	result := make([]*generated.DatabaseCluster, len(domains))
	for i, domain := range domains {
		result[i] = m.ToGraphQLCluster(domain)
	}
	return result
}

// toGraphQLStatus converts a domain ClusterStatus to its GraphQL representation.
func (m *ClusterMapper) toGraphQLStatus(status domainCluster.ClusterStatus) generated.ClusterStatus {
	switch status {
	case domainCluster.ClusterStatusActive:
		return generated.ClusterStatusActive
	case domainCluster.ClusterStatusDisabled:
		return generated.ClusterStatusDisabled
	default:
		return generated.ClusterStatusActive
	}
}

// ToUpdateProjectClusterCommand converts a GraphQL UpdateClusterConnectionInput to an application command.
func (m *ClusterMapper) ToUpdateProjectClusterCommand(
	orgName, projectSlug string,
	input generated.UpdateClusterConnectionInput,
) clusterApp.UpdateProjectClusterCommand {
	cmd := clusterApp.UpdateProjectClusterCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	}

	if input.Title != nil {
		cmd.Title = input.Title
	}
	if input.Description != nil {
		cmd.Description = input.Description
	}
	if input.ConnectionInfo != nil {
		ci := input.ConnectionInfo
		cmd.Host = bizutils.StringPtr(ci.Host)
		cmd.Port = bizutils.IntPtr(int(ci.Port))
		cmd.Username = bizutils.StringPtr(ci.Username)
		cmd.Password = bizutils.StringPtr(ci.Password)
		if ci.ConnectionTimeout != nil {
			cmd.ConnectionTimeout = bizutils.IntPtr(int(*ci.ConnectionTimeout))
		}
	}
	if input.SkipConnectionTest != nil {
		cmd.SkipConnectionTest = *input.SkipConnectionTest
	}

	return cmd
}
