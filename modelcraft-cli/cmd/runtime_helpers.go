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

type runtimeCreds struct {
	Server      string
	OrgName     string
	AccessToken string
}

// newGraphQLClient builds a GraphQLClient reading impersonation flags from
// the command tree. Persistent flags --as-userid / --as-username / --as-role
// are defined on the root command and inherited by all subcommands.
func newGraphQLClient(cmd *cobra.Command) client.GraphQLClient {
	return client.GraphQLClient{
		HTTPClient: http.DefaultClient,
		Impersonation: client.Impersonation{
			UserID:   flagOrEmpty(cmd, "as-userid"),
			UserName: flagOrEmpty(cmd, "as-username"),
			Roles:    flagOrEmpty(cmd, "as-role"),
		},
	}
}

func flagOrEmpty(cmd *cobra.Command, name string) string {
	v, _ := cmd.Flags().GetString(name)
	return v
}

func resolveRuntimeContext(_ *cobra.Command, credentialsPath, project, rawPath string) (map[string]any, resource.ModelPath, runtimeCreds, error) {
	creds, err := loadCredentials(credentialsPath)
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

func bindRuntimeSharedFlags(cmd *cobra.Command, credentialsPath, project *string) {
	cmd.Flags().StringVar(credentialsPath, "credentials", config.DefaultPath(), "Credential file path")
	cmd.Flags().StringVar(project, "project", "", "Project slug")
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
