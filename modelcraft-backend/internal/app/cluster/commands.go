package cluster

// CreateClusterCommand represents a command to create a database cluster.
// Used internally when creating a project (atomic project+cluster creation).
type CreateClusterCommand struct {
	OrgName           string
	ProjectSlug       string
	Title             string
	Description       string
	Host              string
	Port              int
	Username          string
	Password          string
	ConnectionTimeout int
}

// UpdateProjectClusterCommand represents a command to update the cluster connection for a project.
// Since each project has exactly one cluster, the cluster is identified by project alone.
type UpdateProjectClusterCommand struct {
	OrgName            string
	ProjectSlug        string
	Title              *string
	Description        *string
	Host               *string
	Port               *int
	Username           *string
	Password           *string
	ConnectionTimeout  *int
	SkipConnectionTest bool
}

// TestConnectionCommand represents a command to test database connection.
type TestConnectionCommand struct {
	Host              string
	Port              int
	Username          string
	Password          string
	ConnectionTimeout *int
}
