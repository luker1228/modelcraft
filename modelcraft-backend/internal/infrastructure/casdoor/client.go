package casdoor

import (
	"context"
	"fmt"
	"modelcraft/pkg/logfacade"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

// Client wraps the Casdoor SDK client for organization and user management.
type Client struct {
	sdk    *casdoorsdk.Client
	logger logfacade.Logger
}

// Config holds the configuration for Casdoor client.
type Config struct {
	Endpoint     string // Casdoor server URL (e.g., http://localhost:8000)
	ClientID     string // Application client ID
	ClientSecret string // Application client secret
	Certificate  string // X.509 certificate in PEM format
	Organization string // Default organization name
	Application  string // Application name
}

// NewClient creates a new Casdoor client wrapper.
func NewClient(config *Config, logger logfacade.Logger) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate required configuration
	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	if config.Certificate == "" {
		return nil, fmt.Errorf("certificate is required")
	}

	// Create Casdoor SDK client
	sdk := casdoorsdk.NewClient(
		config.Endpoint,
		config.ClientID,
		config.ClientSecret,
		config.Certificate,
		config.Organization,
		config.Application,
	)

	logger.Infof(context.Background(), "Casdoor client initialized: endpoint=%s, organization=%s, application=%s",
		config.Endpoint, config.Organization, config.Application)

	return &Client{
		sdk:    sdk,
		logger: logger,
	}, nil
}

// CreateOrganization creates a new organization in Casdoor.
func (c *Client) CreateOrganization(orgName, displayName string) error {
	c.logger.Infof(
		context.Background(),
		"Creating organization in Casdoor: name=%s, displayName=%s",
		orgName, displayName,
	)

	org := &casdoorsdk.Organization{
		Owner:              "admin", // Organizations are owned by admin
		Name:               orgName,
		CreatedTime:        "", // Casdoor will set this
		DisplayName:        displayName,
		WebsiteUrl:         "",
		Favicon:            "",
		PasswordType:       "plain",
		PasswordSalt:       "",
		DefaultAvatar:      "",
		MasterPassword:     "",
		EnableSoftDeletion: false,
	}

	affected, err := c.sdk.AddOrganization(org)
	if err != nil {
		c.logger.Errorf(context.Background(), "Failed to create organization: %v", err)
		return fmt.Errorf("failed to create organization: %w", err)
	}

	if !affected {
		c.logger.Warnf(context.Background(), "Organization creation returned false (may already exist): %s", orgName)
		return fmt.Errorf("organization creation failed or already exists: %s", orgName)
	}

	c.logger.Infof(context.Background(), "Organization created successfully: %s", orgName)
	return nil
}

// UpdateUserOrganization updates the user's owner (organization) field.
func (c *Client) UpdateUserOrganization(userName, currentOrg, newOrg string) error {
	c.logger.Infof(context.Background(), "Updating user organization: user=%s, currentOrg=%s, newOrg=%s",
		userName, currentOrg, newOrg)

	// Get the user first
	user, err := c.sdk.GetUser(userName)
	if err != nil {
		c.logger.Errorf(context.Background(), "Failed to get user: %v", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %s", userName)
	}

	// Update the owner field
	user.Owner = newOrg

	// Update the user
	affected, err := c.sdk.UpdateUser(user)
	if err != nil {
		c.logger.Errorf(context.Background(), "Failed to update user organization: %v", err)
		return fmt.Errorf("failed to update user organization: %w", err)
	}

	if !affected {
		c.logger.Warnf(context.Background(), "User update returned false: %s", userName)
		return fmt.Errorf("user update failed: %s", userName)
	}

	c.logger.Infof(context.Background(), "User organization updated successfully: %s -> %s", userName, newOrg)
	return nil
}

// GetOrganization retrieves an organization by name.
func (c *Client) GetOrganization(orgName string) (*casdoorsdk.Organization, error) {
	org, err := c.sdk.GetOrganization(orgName)
	if err != nil {
		c.logger.Errorf(context.Background(), "Failed to get organization: %v", err)
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

// GetUser retrieves a user by name.
func (c *Client) GetUser(userName string) (*casdoorsdk.User, error) {
	user, err := c.sdk.GetUser(userName)
	if err != nil {
		c.logger.Errorf(context.Background(), "Failed to get user: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}
