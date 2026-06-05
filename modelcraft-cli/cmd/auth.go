package cmd

import (
	"errors"
	"net/http"
	"os"

	"modelcraft-cli/internal/app"
	"modelcraft-cli/internal/client"
	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func newAuthCommand() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage end-user authentication",
		Example: "  mc auth login --token mc_pat_xxx\n" +
			"  mc auth status\n" +
			"  mc auth switch-project sales\n" +
			"  mc auth logout",
	}
	authCmd.AddCommand(newAuthLoginCommand())
	authCmd.AddCommand(newAuthLogoutCommand())
	authCmd.AddCommand(newAuthStatusCommand())
	authCmd.AddCommand(newAuthSwitchProjectCommand())
	return authCmd
}

const defaultServer = "http://lukemxjia.devcloud.woa.com:9080"

func newAuthLoginCommand() *cobra.Command {
	var server, token, credentialsPath string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with a Personal Access Token (PAT)",
		Example: "  mc auth login --token mc_pat_xxx\n" +
			"  mc auth login --token mc_pat_xxx --server http://gateway:9080",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return output.NewCLIError("MISSING_REQUIRED_FLAG", "Missing required flag.", true,
					"Run 'mc auth login --help' to inspect required flags.",
					map[string]any{"flag": "token"})
			}

			authClient := client.AuthClient{HTTPClient: http.DefaultClient}
			creds, err := authClient.Whoami(cmd.Context(), server, token)
			if err != nil {
				return err
			}
			if err := config.Save(credentialsPath, *creds); err != nil {
				return output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"server":         creds.Server,
				"orgName":        creds.OrgName,
				"userId":         creds.UserID,
				"currentProject": creds.CurrentProject,
			}, nil)
		},
	}

	cmd.Flags().StringVar(&server, "server", defaultServer, "Gateway base URL")
	cmd.Flags().StringVar(&token, "token", "", "Personal Access Token (mc_pat_xxx)")
	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newAuthLogoutCommand() *cobra.Command {
	var credentialsPath string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear local credentials",
		Example: "  mc auth logout\n" +
			"  mc auth logout --credentials /tmp/mc-credentials.json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.Remove(credentialsPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
				}
				return output.NewCLIError("IO_ERROR", "Failed to remove credential file.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"loggedOut": true}, nil)
		},
	}

	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newAuthStatusCommand() *cobra.Command {
	var credentialsPath string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		Example: "  mc auth status\n" +
			"  mc auth status --credentials /tmp/mc-credentials.json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.Load(credentialsPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
				}
				return err
			}

			resolved, err := app.ResolveContext(creds, "")
			if err != nil {
				return err
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"server":         resolved.Server,
				"orgName":        resolved.OrgName,
				"userId":         resolved.UserID,
				"currentProject": resolved.CurrentProject,
			}, nil)
		},
	}

	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newAuthSwitchProjectCommand() *cobra.Command {
	var credentialsPath string

	cmd := &cobra.Command{
		Use:   "switch-project <slug>",
		Short: "Set local default project context",
		Example: "  mc auth switch-project sales\n" +
			"  mc auth switch-project analytics --credentials /tmp/mc-credentials.json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			creds, err := loadCredentials(credentialsPath)
			if err != nil {
				return err
			}

			// Validate by fetching accessible projects from backend in real-time.
			projects, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).CatalogProjects(
				cmd.Context(),
				creds.Server,
				creds.OrgName,
				creds.AccessToken,
			)
			if err != nil {
				return err
			}
			found := false
			for _, p := range projects {
				if p.Slug == slug {
					found = true
					break
				}
			}
			if !found {
				return output.NewCLIError(
					"PROJECT_NOT_FOUND",
					"Project is not accessible for the current user.",
					false,
					"Run 'mc catalog projects' to inspect available projects.",
					map[string]any{"project": slug},
				)
			}

			creds.CurrentProject = slug
			if err := config.Save(credentialsPath, creds); err != nil {
				return output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"currentProject": slug,
			}, nil)
		},
	}

	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}
