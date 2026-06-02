package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"modelcraft-cli/internal/client"
	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	var credentialsPath, project string

	cmd := &cobra.Command{
		Use:   "describe <path>",
		Short: "Describe a runtime model via GraphQL introspection",
		Example: "  mc describe sales.crm.users\n" +
			"  mc describe crm.users --project sales",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			endpoint := clientModelEndpoint(creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model)
			query := `{ __schema { types { name kind fields { name type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } args { name type { kind name ofType { kind name ofType { kind name } } } } } inputFields { name type { kind name ofType { kind name ofType { kind name } } } } } } }`

			var raw struct {
				Schema struct {
					Types []map[string]any `json:"types"`
				} `json:"__schema"`
			}
			err = (client.GraphQLClient{HTTPClient: http.DefaultClient}).Execute(
				cmd.Context(),
				endpoint,
				creds.AccessToken,
				query,
				nil,
				&raw,
			)
			if err != nil {
				return err
			}

			// 过滤掉 GraphQL 内置元类型（以 __ 开头）
			types := make([]map[string]any, 0, len(raw.Schema.Types))
			for _, t := range raw.Schema.Types {
				name, _ := t["name"].(string)
				if !strings.HasPrefix(name, "__") {
					types = append(types, t)
				}
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"types": types}, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	return cmd
}

func clientModelEndpoint(server, org, project, db, model string) string {
	return fmt.Sprintf(
		"%s/end-user/graphql/org/%s/project/%s/db/%s/model/%s",
		strings.TrimRight(server, "/"), org, project, db, model,
	)
}
