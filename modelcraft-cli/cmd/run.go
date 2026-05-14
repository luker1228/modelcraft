package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"

	"modelcraft-cli/internal/client"
	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var credentialsPath, project string

	cmd := &cobra.Command{
		Use:   "run <path> [query]",
		Short: "Execute a GraphQL query against a runtime model",
		Long: `Execute a raw GraphQL query against a runtime model endpoint.

The GraphQL query can be supplied as an argument or piped via stdin.

Examples:
  mc run myproject.mydb.users '{ findMany(take: 5) { id name } }'
  echo '{ count }' | mc run myproject.mydb.users
  mc describe myproject.mydb.users   # discover available fields first`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			// Resolve GraphQL body: argument takes precedence, fallback to stdin.
			var gqlBody string
			if len(args) == 2 {
				gqlBody = strings.TrimSpace(args[1])
			} else {
				raw, readErr := io.ReadAll(os.Stdin)
				if readErr != nil {
					return output.NewCLIError("INVALID_ARGUMENT", "Failed to read query from stdin.", false, "Pipe a GraphQL query body or pass it as the second argument.", nil)
				}
				gqlBody = strings.TrimSpace(string(raw))
			}

			if gqlBody == "" {
				return output.NewCLIError("MISSING_REQUIRED_FLAG", "No GraphQL query provided.", true, "Pass a query as the second argument or pipe it via stdin.", nil)
			}

			result, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Run(
				cmd.Context(), creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model, creds.AccessToken, gqlBody,
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, result, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	return cmd
}
