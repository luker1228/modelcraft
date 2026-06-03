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

func newAuthCommand() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage end-user authentication",
		Example: "  mc auth login --username alice --password '***'\n" +
			"  mc auth status\n" +
			"  mc auth switch-project sales\n" +
			"  mc auth refresh\n" +
			"  mc auth logout",
	}
	authCmd.AddCommand(newAuthLoginCommand())
	authCmd.AddCommand(newAuthLogoutCommand())
	authCmd.AddCommand(newAuthRefreshCommand())
	authCmd.AddCommand(newAuthStatusCommand())
	authCmd.AddCommand(newAuthSwitchProjectCommand())
	return authCmd
}

const defaultServer = "http://lukemxjia.devcloud.woa.com:9080"

func newAuthLoginCommand() *cobra.Command {
	var server, org, username, password, token, credentialsPath string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with end-user credentials or a PAT token",
		Example: "  mc auth login --token mc_pat_xxx\n" +
			"  mc auth login --username alice --password '***'\n" +
			"  mc auth login --username alice --password '***' --server http://lukemxjia.devcloud.woa.com:9080",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			authClient := client.AuthClient{HTTPClient: http.DefaultClient}

			var creds *config.Credentials
			var err error

			if token != "" {
				// PAT-based login: call whoami to resolve identity and projects.
				creds, err = authClient.Whoami(cmd.Context(), server, token)
			} else {
				// Username/password login — validate required flags before network call.
				if username == "" {
					return output.NewCLIError("MISSING_REQUIRED_FLAG", "Missing required flag.", true,
						"Run 'mc auth login --help' to inspect required flags.",
						map[string]any{"flag": "username"})
				}
				if password == "" {
					return output.NewCLIError("MISSING_REQUIRED_FLAG", "Missing required flag.", true,
						"Run 'mc auth login --help' to inspect required flags.",
						map[string]any{"flag": "password"})
				}
				creds, err = authClient.Login(cmd.Context(), server, org, username, password)
			}
			if err != nil {
				return err
			}
			if err := config.Save(credentialsPath, *creds); err != nil {
				return output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"server":   creds.Server,
				"orgName":  creds.OrgName,
				"userId":   creds.UserID,
				"projects": creds.Projects,
			}, nil)
		},
	}

	cmd.Flags().StringVar(&server, "server", defaultServer, "Gateway base URL")
	cmd.Flags().StringVar(&token, "token", "", "Personal Access Token (mc_pat_xxx) — skips username/password")
	cmd.Flags().StringVar(&org, "org", "", "Organization slug (optional, auto-resolved by username)")
	cmd.Flags().StringVar(&username, "username", "", "End-user username")
	cmd.Flags().StringVar(&password, "password", "", "End-user password")
	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newAuthLogoutCommand() *cobra.Command {
	var credentialsPath string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear local credentials",
		Example: "  mc auth logout\n" +
			"  mc auth logout --credentials /tmp/mc-credentials.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.Load(credentialsPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
				}
				return err
			}

			authClient := client.AuthClient{HTTPClient: http.DefaultClient}
			if err := authClient.Logout(cmd.Context(), creds.Server, creds.OrgName, creds.RefreshToken); err != nil {
				return err
			}
			if err := os.Remove(credentialsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return output.NewCLIError("IO_ERROR", "Failed to remove credential file.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"loggedOut": true}, nil)
		},
	}

	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}

func newAuthRefreshCommand() *cobra.Command {
	var credentialsPath string

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh access token using local refresh token",
		Example: "  mc auth refresh\n" +
			"  mc auth refresh --credentials /tmp/mc-credentials.json",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := config.Load(credentialsPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
				}
				return err
			}

			authClient := client.AuthClient{HTTPClient: http.DefaultClient}
			mgr := authsession.Manager{}
			fresh, err := mgr.EnsureFresh(cmd.Context(), creds, authClient)
			if err != nil {
				return err
			}
			if err := config.Save(credentialsPath, fresh); err != nil {
				return output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"server":         fresh.Server,
				"orgName":        fresh.OrgName,
				"userId":         fresh.UserID,
				"currentProject": fresh.CurrentProject,
			}, nil)
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
		Args:  cobra.NoArgs,
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
				"projects":       resolved.Projects,
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			creds, err := config.Load(credentialsPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return output.NewCLIError("UNAUTHENTICATED", "No local session found.", true, "Run 'mc auth login'.", nil)
				}
				return err
			}

			updated, err := authsession.SwitchProject(creds, slug)
			if err != nil {
				return err
			}
			if err := config.Save(credentialsPath, updated); err != nil {
				return output.NewCLIError("IO_ERROR", "Failed to persist credentials.", false, "Check filesystem permissions and retry.", map[string]any{"path": credentialsPath})
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
				"currentProject": updated.CurrentProject,
			}, nil)
		},
	}

	cmd.Flags().StringVar(&credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	return cmd
}
