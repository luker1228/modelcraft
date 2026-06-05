package cmd

import (
	"errors"
	"net/http"
	"os"

	"modelcraft-cli/internal/app"
	authsession "modelcraft-cli/internal/auth"
	"modelcraft-cli/internal/client"
	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func newCatalogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Discover accessible projects, databases, and models",
		Example: "  mc catalog projects\n" +
			"  mc catalog databases --project sales\n" +
			"  mc catalog models --project sales --database crm",
	}
	cmd.AddCommand(newCatalogProjectsCommand())
	cmd.AddCommand(newCatalogDatabasesCommand())
	cmd.AddCommand(newCatalogModelsCommand())
	return cmd
}

func newCatalogProjectsCommand() *cobra.Command {
	var credentialsPath string
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List accessible projects",
		Example: "  mc catalog projects\n" +
			"  mc catalog projects --credentials /tmp/mc-credentials.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := loadCredentials(credentialsPath)
			if err != nil {
				return err
			}
			items, err := (client.GraphQLClient{}).CatalogProjects(cmd.Context(), creds)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"items": items}, nil)
		},
	}
	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newCatalogDatabasesCommand() *cobra.Command {
	var credentialsPath, project string
	cmd := &cobra.Command{
		Use:   "databases",
		Short: "List databases in a project",
		Example: "  mc catalog databases --project sales\n" +
			"  mc catalog databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := loadFreshCredentials(cmd, credentialsPath)
			if err != nil {
				return err
			}

			resolved, err := app.ResolveContext(creds, project)
			if err != nil {
				return err
			}
			if resolved.CurrentProject == "" {
				return output.NewCLIError("NO_PROJECT_CONTEXT", "No project context is selected.", true, "Use --project <slug> or run 'mc auth switch-project <slug>'.", nil)
			}

			// Validate that the requested project is accessible.
			if err := validateProjectAccess(resolved); err != nil {
				return err
			}

			items, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).CatalogDatabases(
				cmd.Context(),
				resolved.Server,
				resolved.OrgName,
				resolved.CurrentProject,
				resolved.AccessToken,
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"items": items}, map[string]any{"project": resolved.CurrentProject})
		},
	}
	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	cmd.Flags().StringVar(&project, "project", "", "Project slug")
	return cmd
}

func newCatalogModelsCommand() *cobra.Command {
	var credentialsPath, project, database string
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List models in a database",
		Example: "  mc catalog models --database crm --project sales\n" +
			"  mc catalog models --database crm",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := loadFreshCredentials(cmd, credentialsPath)
			if err != nil {
				return err
			}

			resolved, err := app.ResolveContext(creds, project)
			if err != nil {
				return err
			}
			if resolved.CurrentProject == "" {
				return output.NewCLIError("NO_PROJECT_CONTEXT", "No project context is selected.", true, "Use --project <slug> or run 'mc auth switch-project <slug>'.", nil)
			}
			if database == "" {
				return output.NewCLIError("MISSING_REQUIRED_FLAG", "Missing required flag.", true, "Run 'mc catalog models --help' to inspect required flags.", map[string]any{"flag": "database"})
			}

			// Validate that the requested project is accessible.
			if err := validateProjectAccess(resolved); err != nil {
				return err
			}

			items, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).CatalogModels(
				cmd.Context(),
				resolved.Server,
				resolved.OrgName,
				resolved.CurrentProject,
				database,
				resolved.AccessToken,
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"items": items}, map[string]any{"project": resolved.CurrentProject, "database": database})
		},
	}
	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	cmd.Flags().StringVar(&project, "project", "", "Project slug")
	cmd.Flags().StringVar(&database, "database", "", "Database name")
	return cmd
}

// validateProjectAccess checks that resolved.CurrentProject is in the user's accessible project list.
// Returns PROJECT_NOT_FOUND if the project is not accessible.
func validateProjectAccess(creds config.Credentials) error {
	if len(creds.Projects) == 0 {
		// No project list cached (e.g. username/password login without whoami) — skip check.
		return nil
	}
	for _, p := range creds.Projects {
		if p.Slug == creds.CurrentProject {
			return nil
		}
	}
	return output.NewCLIError(
		"PROJECT_NOT_FOUND",
		"Project is not accessible for the current user.",
		false,
		"Run 'mc catalog projects' to inspect available projects.",
		map[string]any{"project": creds.CurrentProject},
	)
}

func loadCredentials(credentialsPath string) (config.Credentials, error) {
	creds, err := config.Load(credentialsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config.Credentials{}, output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
		}
		return config.Credentials{}, err
	}
	return creds, nil
}

func loadFreshCredentials(cmd *cobra.Command, credentialsPath string) (config.Credentials, error) {
	creds, err := loadCredentials(credentialsPath)
	if err != nil {
		return config.Credentials{}, err
	}

	mgr := authsession.Manager{}
	authClient := client.AuthClient{HTTPClient: http.DefaultClient}
	fresh, err := mgr.EnsureFresh(cmd.Context(), creds, authClient)
	if err != nil {
		return config.Credentials{}, err
	}
	if err := config.Save(credentialsPath, fresh); err != nil {
		return config.Credentials{}, output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
	}
	return fresh, nil
}
