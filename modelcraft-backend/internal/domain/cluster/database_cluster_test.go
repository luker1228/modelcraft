package cluster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseCluster(t *testing.T) {
	setupTestCipher(t)

	t.Run("default connection timeout should be 5 seconds", func(t *testing.T) {
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"Test Cluster",
			"localhost",
			3306,
			"root",
			"password",
		)

		assert.NoError(t, err)
		assert.NotNil(t, cluster)
		assert.Equal(t, 5, cluster.ConnectionTimeout)
	})

	t.Run("should fail with empty password", func(t *testing.T) {
		// Given: an empty password
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"Test Cluster",
			"localhost",
			3306,
			"root",
			"",
		)

		// Then: should return error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "plain password cannot be empty")
	})

	t.Run("should fail with invalid org name", func(t *testing.T) {
		// Given: an empty org name
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"",
			"test-project",
			"Test Cluster",
			"localhost",
			3306,
			"root",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "org name cannot be empty")
	})

	t.Run("should fail with invalid project slug", func(t *testing.T) {
		// Given: an empty project slug
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"",
			"Test Cluster",
			"localhost",
			3306,
			"root",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "project slug cannot be empty")
	})

	t.Run("should fail with invalid title", func(t *testing.T) {
		// Given: an empty title
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"",
			"localhost",
			3306,
			"root",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "cluster title cannot be empty")
	})

	t.Run("should fail with invalid host", func(t *testing.T) {
		// Given: an empty host
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"Test Cluster",
			"",
			3306,
			"root",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "host cannot be empty")
	})

	t.Run("should fail with invalid port", func(t *testing.T) {
		// Given: an invalid port (-1)
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"Test Cluster",
			"localhost",
			-1,
			"root",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "port must be between 1 and 65535")
	})

	t.Run("should fail with invalid username", func(t *testing.T) {
		// Given: an empty username
		// When: creating a new database cluster
		cluster, err := NewDatabaseCluster(
			"built-in",
			"test-project",
			"Test Cluster",
			"localhost",
			3306,
			"",
			"password",
		)

		// Then: should return validation error
		assert.Error(t, err)
		assert.Nil(t, cluster)
		assert.Contains(t, err.Error(), "username cannot be empty")
	})
}

func TestDatabaseCluster_Validate(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("password")

	tests := []struct {
		name        string
		cluster     *DatabaseCluster
		wantErr     bool
		errContains string
	}{
		{
			name: "valid cluster with default timeout",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr: false,
		},
		{
			name: "invalid cluster with empty org name",
			cluster: &DatabaseCluster{
				OrgName:           "",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "org name cannot be empty",
		},
		{
			name: "invalid cluster with empty project slug",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "project slug cannot be empty",
		},
		{
			name: "invalid cluster with empty title",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "cluster title cannot be empty",
		},
		{
			name: "invalid cluster with empty host",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "host cannot be empty",
		},
		{
			name: "invalid cluster with zero port",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              0,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "port must be between 1 and 65535",
		},
		{
			name: "invalid cluster with negative port",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              -1,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "port must be between 1 and 65535",
		},
		{
			name: "invalid cluster with port exceeding maximum",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              65536,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "port must be between 1 and 65535",
		},
		{
			name: "invalid cluster with empty username",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "username cannot be empty",
		},
		{
			name: "valid cluster with timeout at minimum boundary (5 seconds)",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			wantErr: false,
		},
		{
			name: "valid cluster with timeout at maximum boundary (15 seconds)",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 15,
				Status:            ClusterStatusActive,
			},
			wantErr: false,
		},
		{
			name: "invalid cluster with timeout below minimum",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 4,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "connection timeout must be between 5 and 15 seconds",
		},
		{
			name: "invalid cluster with timeout above maximum",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 16,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "connection timeout must be between 5 and 15 seconds",
		},
		{
			name: "invalid cluster with zero timeout",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 0,
				Status:            ClusterStatusActive,
			},
			wantErr:     true,
			errContains: "connection timeout must be between 5 and 15 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cluster.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseCluster_UpdateConnectionTimeout(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("password")

	t.Run("update connection timeout to valid value", func(t *testing.T) {
		cluster := &DatabaseCluster{
			OrgName:           "built-in",
			ProjectSlug:       "test-project",
			Title:             "Test Cluster",
			Host:              "localhost",
			Port:              3306,
			Username:          "root",
			Password:          *password,
			ConnectionTimeout: 5,
			Status:            ClusterStatusActive,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		oldUpdatedAt := cluster.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		err := cluster.UpdateConnectionTimeout(10)
		assert.NoError(t, err)
		assert.Equal(t, 10, cluster.ConnectionTimeout)
		assert.True(t, cluster.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("update connection timeout to invalid value should fail and rollback", func(t *testing.T) {
		cluster := &DatabaseCluster{
			OrgName:           "built-in",
			ProjectSlug:       "test-project",
			Title:             "Test Cluster",
			Host:              "localhost",
			Port:              3306,
			Username:          "root",
			Password:          *password,
			ConnectionTimeout: 5,
			Status:            ClusterStatusActive,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := cluster.UpdateConnectionTimeout(20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection timeout must be between 5 and 15 seconds")
		assert.Equal(t, 5, cluster.ConnectionTimeout) // Should remain unchanged
	})

	t.Run("update connection timeout to minimum boundary", func(t *testing.T) {
		cluster := &DatabaseCluster{
			OrgName:           "built-in",
			ProjectSlug:       "test-project",
			Title:             "Test Cluster",
			Host:              "localhost",
			Port:              3306,
			Username:          "root",
			Password:          *password,
			ConnectionTimeout: 10,
			Status:            ClusterStatusActive,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := cluster.UpdateConnectionTimeout(5)
		assert.NoError(t, err)
		assert.Equal(t, 5, cluster.ConnectionTimeout)
	})

	t.Run("update connection timeout to maximum boundary", func(t *testing.T) {
		cluster := &DatabaseCluster{
			OrgName:           "built-in",
			ProjectSlug:       "test-project",
			Title:             "Test Cluster",
			Host:              "localhost",
			Port:              3306,
			Username:          "root",
			Password:          *password,
			ConnectionTimeout: 5,
			Status:            ClusterStatusActive,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		err := cluster.UpdateConnectionTimeout(15)
		assert.NoError(t, err)
		assert.Equal(t, 15, cluster.ConnectionTimeout)
	})
}

func TestClusterLocator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		locator     ClusterLocator
		wantErr     bool
		errContains string
	}{
		{
			name: "valid cluster locator with both fields",
			locator: ClusterLocator{
				OrgName:     "built-in",
				ProjectSlug: "test_project",
			},
			wantErr: false,
		},
		{
			name: "valid cluster locator with special characters",
			locator: ClusterLocator{
				OrgName:     "my-org",
				ProjectSlug: "my_project_123",
			},
			wantErr: false,
		},
		{
			name: "invalid cluster locator with empty org name",
			locator: ClusterLocator{
				OrgName:     "",
				ProjectSlug: "test_project",
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "invalid cluster locator with empty project slug",
			locator: ClusterLocator{
				OrgName:     "built-in",
				ProjectSlug: "",
			},
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name: "invalid cluster locator with both fields empty",
			locator: ClusterLocator{
				OrgName:     "",
				ProjectSlug: "",
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a cluster locator with specific values
			// When: validating the cluster locator
			err := tt.locator.Validate()

			// Then: verify error expectation
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewClusterLocator(t *testing.T) {
	tests := []struct {
		name        string
		orgName     string
		projectSlug string
		wantErr     bool
		errContains string
	}{
		{
			name:        "create valid cluster locator",
			orgName:     "built-in",
			projectSlug: "test_project",
			wantErr:     false,
		},
		{
			name:        "create cluster locator with special characters",
			orgName:     "my-org-123",
			projectSlug: "project_v2",
			wantErr:     false,
		},
		{
			name:        "fail to create cluster locator with empty org name",
			orgName:     "",
			projectSlug: "test_project",
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name:        "fail to create cluster locator with empty project slug",
			orgName:     "built-in",
			projectSlug: "",
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name:        "fail to create cluster locator with both fields empty",
			orgName:     "",
			projectSlug: "",
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: org name and project slug parameters
			// When: creating a new cluster locator
			locator, err := NewClusterLocator(tt.orgName, tt.projectSlug)

			// Then: verify creation result
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.orgName, locator.OrgName)
				assert.Equal(t, tt.projectSlug, locator.ProjectSlug)
			}
		})
	}
}

func TestClusterLocator_GetFullPath(t *testing.T) {
	tests := []struct {
		name         string
		locator      ClusterLocator
		expectedPath string
	}{
		{
			name: "get full path with standard values",
			locator: ClusterLocator{
				OrgName:     "built-in",
				ProjectSlug: "test_project",
			},
			expectedPath: "built-in.test_project",
		},
		{
			name: "get full path with special characters",
			locator: ClusterLocator{
				OrgName:     "my-org",
				ProjectSlug: "my_project_123",
			},
			expectedPath: "my-org.my_project_123",
		},
		{
			name: "get full path with single character values",
			locator: ClusterLocator{
				OrgName:     "o",
				ProjectSlug: "p",
			},
			expectedPath: "o.p",
		},
		{
			name: "get full path with long values",
			locator: ClusterLocator{
				OrgName:     "very-long-organization-name-12345",
				ProjectSlug: "very_long_project_slug_67890",
			},
			expectedPath: "very-long-organization-name-12345.very_long_project_slug_67890",
		},
		{
			name: "get full path with empty values returns dot only",
			locator: ClusterLocator{
				OrgName:     "",
				ProjectSlug: "",
			},
			expectedPath: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a cluster locator with org name and project slug
			// When: calling GetFullPath
			fullPath := tt.locator.GetFullPath()

			// Then: verify the full path format is correct
			assert.Equal(t, tt.expectedPath, fullPath)
		})
	}
}

func TestDatabaseCluster_GetClusterLocator(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("password")

	tests := []struct {
		name            string
		cluster         *DatabaseCluster
		expectedOrgName string
		expectedSlug    string
	}{
		{
			name: "get cluster locator from valid cluster",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test_project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			expectedOrgName: "built-in",
			expectedSlug:    "test_project",
		},
		{
			name: "get cluster locator with special characters",
			cluster: &DatabaseCluster{
				OrgName:           "my-org",
				ProjectSlug:       "my_project_123",
				Title:             "Production Cluster",
				Host:              "prod.example.com",
				Port:              3306,
				Username:          "admin",
				Password:          *password,
				ConnectionTimeout: 10,
				Status:            ClusterStatusActive,
			},
			expectedOrgName: "my-org",
			expectedSlug:    "my_project_123",
		},
		{
			name: "get cluster locator with minimal cluster data",
			cluster: &DatabaseCluster{
				OrgName:     "org",
				ProjectSlug: "proj",
			},
			expectedOrgName: "org",
			expectedSlug:    "proj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a database cluster with org name and project slug
			// When: calling GetClusterLocator
			locator := tt.cluster.GetClusterLocator()

			// Then: verify the locator contains correct values
			assert.NotNil(t, locator)
			assert.Equal(t, tt.expectedOrgName, locator.OrgName)
			assert.Equal(t, tt.expectedSlug, locator.ProjectSlug)
		})
	}
}

func TestDatabaseCluster_IsActive(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("password")

	tests := []struct {
		name           string
		clusterStatus  ClusterStatus
		expectedActive bool
	}{
		{
			name:           "active cluster returns true",
			clusterStatus:  ClusterStatusActive,
			expectedActive: true,
		},
		{
			name:           "disabled cluster returns false",
			clusterStatus:  ClusterStatusDisabled,
			expectedActive: false,
		},
		{
			name:           "empty status returns false",
			clusterStatus:  "",
			expectedActive: false,
		},
		{
			name:           "invalid status returns false",
			clusterStatus:  "invalid",
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a database cluster with a specific status
			cluster := &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test_project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            tt.clusterStatus,
			}

			// When: checking if cluster is active
			isActive := cluster.IsActive()

			// Then: verify the active status matches expectation
			assert.Equal(t, tt.expectedActive, isActive)
		})
	}
}

func TestDatabaseCluster_SetDefaults(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("password")

	tests := []struct {
		name                      string
		cluster                   *DatabaseCluster
		expectedPort              int
		expectedConnectionTimeout int
		expectedStatus            ClusterStatus
		expectedVersion           int64
	}{
		{
			name: "set default port when zero",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              0,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 10,
				Status:            ClusterStatusActive,
				Version:           1,
			},
			expectedPort:              3306,
			expectedConnectionTimeout: 10,
			expectedStatus:            ClusterStatusActive,
			expectedVersion:           1,
		},
		{
			name: "set default connection timeout when zero",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3307,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 0,
				Status:            ClusterStatusActive,
				Version:           1,
			},
			expectedPort:              3307,
			expectedConnectionTimeout: 5,
			expectedStatus:            ClusterStatusActive,
			expectedVersion:           1,
		},
		{
			name: "set default status when empty",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 10,
				Status:            "",
				Version:           1,
			},
			expectedPort:              3306,
			expectedConnectionTimeout: 10,
			expectedStatus:            ClusterStatusActive,
			expectedVersion:           1,
		},
		{
			name: "set default version when zero",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 10,
				Status:            ClusterStatusActive,
				Version:           0,
			},
			expectedPort:              3306,
			expectedConnectionTimeout: 10,
			expectedStatus:            ClusterStatusActive,
			expectedVersion:           1,
		},
		{
			name: "set all defaults when all zero or empty",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              0,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 0,
				Status:            "",
				Version:           0,
			},
			expectedPort:              3306,
			expectedConnectionTimeout: 5,
			expectedStatus:            ClusterStatusActive,
			expectedVersion:           1,
		},
		{
			name: "preserve non-zero values",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test-project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3307,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 15,
				Status:            ClusterStatusDisabled,
				Version:           5,
			},
			expectedPort:              3307,
			expectedConnectionTimeout: 15,
			expectedStatus:            ClusterStatusDisabled,
			expectedVersion:           5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a database cluster with specific default values
			// When: calling SetDefaults
			tt.cluster.SetDefaults()

			// Then: verify defaults are set correctly
			assert.Equal(t, tt.expectedPort, tt.cluster.Port)
			assert.Equal(t, tt.expectedConnectionTimeout, tt.cluster.ConnectionTimeout)
			assert.Equal(t, tt.expectedStatus, tt.cluster.Status)
			assert.Equal(t, tt.expectedVersion, tt.cluster.Version)
		})
	}
}

func TestDatabaseCluster_GetConnectionInfo(t *testing.T) {
	setupTestCipher(t)
	password, _ := NewByPlain("mySecurePassword")

	tests := []struct {
		name                      string
		cluster                   *DatabaseCluster
		expectedHost              string
		expectedPort              int
		expectedUsername          string
		expectedConnectionTimeout int
	}{
		{
			name: "get connection info from standard cluster",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test_project",
				Title:             "Test Cluster",
				Host:              "localhost",
				Port:              3306,
				Username:          "root",
				Password:          *password,
				ConnectionTimeout: 5,
				Status:            ClusterStatusActive,
			},
			expectedHost:              "localhost",
			expectedPort:              3306,
			expectedUsername:          "root",
			expectedConnectionTimeout: 5,
		},
		{
			name: "get connection info from production cluster",
			cluster: &DatabaseCluster{
				OrgName:           "my-org",
				ProjectSlug:       "prod_project",
				Title:             "Production Cluster",
				Host:              "prod-db.example.com",
				Port:              3307,
				Username:          "admin",
				Password:          *password,
				ConnectionTimeout: 15,
				Status:            ClusterStatusActive,
			},
			expectedHost:              "prod-db.example.com",
			expectedPort:              3307,
			expectedUsername:          "admin",
			expectedConnectionTimeout: 15,
		},
		{
			name: "get connection info with IP address host",
			cluster: &DatabaseCluster{
				OrgName:           "built-in",
				ProjectSlug:       "test_project",
				Title:             "Test Cluster",
				Host:              "192.168.1.100",
				Port:              3306,
				Username:          "dbuser",
				Password:          *password,
				ConnectionTimeout: 10,
				Status:            ClusterStatusActive,
			},
			expectedHost:              "192.168.1.100",
			expectedPort:              3306,
			expectedUsername:          "dbuser",
			expectedConnectionTimeout: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a database cluster with connection parameters
			// When: getting connection info
			connInfo := tt.cluster.GetConnectionInfo()

			// Then: verify all connection parameters are correct
			assert.Equal(t, tt.expectedHost, connInfo.Host)
			assert.Equal(t, tt.expectedPort, connInfo.Port)
			assert.Equal(t, tt.expectedUsername, connInfo.Username)
			assert.Equal(t, tt.expectedConnectionTimeout, connInfo.ConnectionTimeout)

			// Verify password is included and can be decrypted
			plainPassword, err := connInfo.Password.GetPlainPassword()
			assert.NoError(t, err)
			assert.Equal(t, "mySecurePassword", plainPassword)
		})
	}
}
