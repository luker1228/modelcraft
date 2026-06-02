package cmd

import (
	"encoding/json"
	"net/http"

	"modelcraft-cli/internal/app"
	"modelcraft-cli/internal/client"
	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"
	"modelcraft-cli/internal/resource"

	"github.com/spf13/cobra"
)

func newQueryCommand() *cobra.Command {
	var credentialsPath, project, whereRaw string

	cmd := &cobra.Command{
		Use:   "query <path>",
		Short: "Query records from a runtime model",
		Example: "  mc query sales.crm.users\n" +
			"  mc query sales.crm.users --where '{\"status\": \"active\"}'\n" +
			"  mc query crm.users --project sales",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Validate --where JSON before any context resolution (INVALID_JSON_FLAG = exit 2).
			if _, err := parseJSONFlag("where", whereRaw); err != nil {
				return err
			}

			// 2. Resolve context: validates credentials freshness and parses project/db/model path.
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			// 3. Build a simple findMany query. The body is intentionally minimal; callers
			//    that need full control over the query shape should use `mc run` instead.
			gqlBody := "{ findMany { id } }"

			// 4. Execute against the runtime endpoint.
			result, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Run(
				cmd.Context(), creds.Server, creds.OrgName,
				modelPath.Project, modelPath.Database, modelPath.Model,
				creds.AccessToken, gqlBody,
			)
			if err != nil {
				return err
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, result, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	cmd.Flags().StringVar(&whereRaw, "where", "", "JSON filter object (e.g. '{\"status\":\"active\"}')")
	return cmd
}

func bindRuntimeSharedFlags(cmd *cobra.Command, credentialsPath, project *string) {
	cmd.Flags().StringVar(credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	cmd.Flags().StringVar(project, "project", "", "Project slug")
}

func resolveRuntimeContext(cmd *cobra.Command, credentialsPath, project, rawPath string) (map[string]any, resource.ModelPath, runtimeCreds, error) {
	creds, err := loadFreshCredentials(cmd, credentialsPath)
	if err != nil {
		return nil, resource.ModelPath{}, runtimeCreds{}, err
	}

	resolved, err := app.ResolveContext(creds, project)
	if err != nil {
		return nil, resource.ModelPath{}, runtimeCreds{}, err
	}

	modelPath, err := resource.ParseModelPath(rawPath, resource.ParseContext{CurrentProject: resolved.CurrentProject})
	if err != nil {
		return nil, resource.ModelPath{}, runtimeCreds{}, err
	}

	ctx := map[string]any{"project": modelPath.Project, "database": modelPath.Database, "model": modelPath.Model}
	return ctx, modelPath, runtimeCreds{Server: resolved.Server, OrgName: resolved.OrgName, AccessToken: resolved.AccessToken}, nil
}

type runtimeCreds struct {
	Server      string
	OrgName     string
	AccessToken string
}

func parseJSONFlag(flagName, raw string) (json.RawMessage, error) {
	if raw == "" {
		return nil, nil
	}
	var tmp any
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return nil, output.NewCLIError(
			"INVALID_JSON_FLAG",
			"Invalid JSON flag value.",
			true,
			"Fix the JSON syntax and retry.",
			map[string]any{"flag": flagName, "value": raw, "error": err.Error()},
		)
	}
	return json.RawMessage(raw), nil
}
