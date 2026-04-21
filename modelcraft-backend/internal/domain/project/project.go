package project

import (
	"fmt"
	"regexp"
	"time"
)

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

const (
	projectSlugMinLen = 3
	// Keep mc_private_{slug} within MySQL database name length (64).
	// len("mc_private_") = 11, so slug max is 53.
	projectSlugMaxLen = 53
)

// Project represents a project entity in the domain
// Uses (OrgName, Name) as composite primary key
type Project struct {
	OrgName     string // Organization name (from Casdoor) - part of primary key
	Slug        string // Project slug (unique within org) - part of primary key
	Title       string // Display title
	Description string
	ClusterID   *string // Cluster ID (nullable, one-to-one relationship)
	Status      ProjectStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var projectSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// isValidProjectSlug validates that a project slug follows the required format:
// - 3-53 characters
// - lowercase letters, digits, and underscores only (no hyphens or special characters)
// - MUST start with a letter
func isValidProjectSlug(name string) bool {
	if len(name) < projectSlugMinLen || len(name) > projectSlugMaxLen {
		return false
	}
	return projectSlugPattern.MatchString(name)
}

// Validate validates the Project entity
func (p *Project) Validate() error {
	if p.OrgName == "" {
		return fmt.Errorf("organization name is required")
	}
	if p.Slug == "" {
		return fmt.Errorf("ProjectSlug cant be blank")
	}
	if !isValidProjectSlug(p.Slug) {
		return fmt.Errorf(
			"project slug MUST be 3-53 characters, lowercase letters/digits/underscores only, " +
				"and start with a letter",
		)
	}
	if p.Title == "" {
		return fmt.Errorf("project title is required")
	}
	if p.Status != ProjectStatusActive && p.Status != ProjectStatusArchived {
		return fmt.Errorf("project status MUST be either 'active' or 'archived'")
	}
	if p.ClusterID != nil && *p.ClusterID == "" {
		return fmt.Errorf("cluster ID cannot be empty if provided")
	}
	return nil
}

// NewProject creates a new Project entity with validation
// Primary key is (orgName, name) composite
func NewProject(orgName, slug, title, description string) (*Project, error) {
	now := time.Now()

	project := &Project{
		OrgName:     orgName,
		Slug:        slug,
		Title:       title,
		Description: description,
		Status:      ProjectStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	return project, nil
}

// UpdateMetadata updates the project's title and description
// Empty string means "do not update this field" for all parameters
func (p *Project) UpdateMetadata(title, description string) error {
	originalTitle := p.Title
	originalDescription := p.Description

	if title != "" {
		p.Title = title
	}
	if description != "" {
		p.Description = description
	}
	p.UpdatedAt = time.Now()

	if err := p.Validate(); err != nil {
		p.Title = originalTitle
		p.Description = originalDescription
		return err
	}

	return nil
}

// Archive marks the project as archived
func (p *Project) Archive() {
	p.Status = ProjectStatusArchived
	p.UpdatedAt = time.Now()
}

// Activate marks the project as active
func (p *Project) Activate() {
	p.Status = ProjectStatusActive
	p.UpdatedAt = time.Now()
}

// IsActive returns true if the project is active
func (p *Project) IsActive() bool {
	return p.Status == ProjectStatusActive
}

// SetCluster sets the cluster ID for the project
func (p *Project) SetCluster(clusterID string) error {
	if clusterID == "" {
		return fmt.Errorf("cluster ID cannot be empty")
	}
	p.ClusterID = &clusterID
	p.UpdatedAt = time.Now()
	return p.Validate()
}

// UnsetCluster removes the cluster association from the project
func (p *Project) UnsetCluster() {
	p.ClusterID = nil
	p.UpdatedAt = time.Now()
}

// GetClusterID returns the cluster ID if set, otherwise returns empty string
func (p *Project) GetClusterID() string {
	if p.ClusterID == nil {
		return ""
	}
	return *p.ClusterID
}

// HasCluster returns true if the project has an associated cluster
func (p *Project) HasCluster() bool {
	return p.ClusterID != nil && *p.ClusterID != ""
}
