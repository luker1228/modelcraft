package adapter

import (
	"modelcraft/internal/domain/project"
	"modelcraft/internal/interfaces/graphql/org/generated"
)

// ProjectMapper provides the singleton project mapper instance.
var ProjectMapper *projectMapper

func init() {
	ProjectMapper = &projectMapper{}
}

type projectMapper struct{}

// ConvertProjectToGraphQL converts domain Project to GraphQL Project
func (p *projectMapper) ConvertProjectToGraphQL(proj *project.Project) *generated.Project {
	return &generated.Project{
		ID:          proj.Slug, // Use slug as ID for backward compatibility
		Slug:        proj.Slug,
		Title:       proj.Title,
		Description: proj.Description,
		Status:      ConvertProjectStatusToGraphQL(proj.Status),
		OrgName:     proj.OrgName,
		// Keep non-null contract for Project.authSchema.
		AuthSchema: &generated.ProjectAuthSchema{Variables: []*generated.AuthVariable{}},
		// Cluster field is resolved separately via ProjectResolver.Cluster()
		CreatedAt: proj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: proj.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ConvertProjectsToGraphQL converts a slice of domain Projects to GraphQL Projects
func (p *projectMapper) ConvertProjectsToGraphQL(projects []*project.Project) []*generated.Project {
	result := make([]*generated.Project, 0, len(projects))
	for _, proj := range projects {
		result = append(result, p.ConvertProjectToGraphQL(proj))
	}
	return result
}

// ConvertProjectStatusToGraphQL converts domain ProjectStatus to GraphQL ProjectStatus
func ConvertProjectStatusToGraphQL(status project.ProjectStatus) generated.ProjectStatus {
	switch status {
	case project.ProjectStatusActive:
		return generated.ProjectStatusActive
	case project.ProjectStatusArchived:
		return generated.ProjectStatusArchived
	default:
		return generated.ProjectStatusActive
	}
}

// ConvertProjectStatusFromGraphQL converts GraphQL ProjectStatus to domain ProjectStatus
func ConvertProjectStatusFromGraphQL(status generated.ProjectStatus) project.ProjectStatus {
	switch status {
	case generated.ProjectStatusActive:
		return project.ProjectStatusActive
	case generated.ProjectStatusArchived:
		return project.ProjectStatusArchived
	default:
		return project.ProjectStatusActive
	}
}
