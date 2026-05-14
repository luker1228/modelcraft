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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			endpoint := clientModelEndpoint(creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model)
			query := `query DescribeModel($typeName: String!) { __type(name: $typeName) { name fields { name type { kind name ofType { kind name ofType { kind name } } } } } }`

			var payload struct {
				Type *struct {
					Name   string `json:"name"`
					Fields []struct {
						Name string `json:"name"`
						Type struct {
							Kind   string `json:"kind"`
							Name   *string `json:"name"`
							OfType *struct {
								Kind   string `json:"kind"`
								Name   *string `json:"name"`
								OfType *struct {
									Kind string  `json:"kind"`
									Name *string `json:"name"`
								} `json:"ofType"`
							} `json:"ofType"`
						} `json:"type"`
					} `json:"fields"`
				} `json:"__type"`
			}
			// Runtime schema prefixes model type names with "T" (e.g. "test005" → "Ttest005")
			err = (client.GraphQLClient{HTTPClient: http.DefaultClient}).Execute(
				cmd.Context(),
				endpoint,
				creds.AccessToken,
				query,
				map[string]any{"typeName": "Query"},
				&payload,
			)
			if err != nil {
				return err
			}
			if payload.Type == nil {
				return output.NewCLIError("MODEL_NOT_FOUND", "Model type was not found by introspection.", false, "Check the model path and retry.", map[string]any{"model": modelPath.Model})
			}

			fields := make([]map[string]any, 0, len(payload.Type.Fields))
			for _, f := range payload.Type.Fields {
				kind, typeName, required, isList := flattenGraphQLType(f.Type)
				fields = append(fields, map[string]any{
					"name":     f.Name,
					"kind":     kind,
					"type":     typeName,
					"required": required,
					"isList":   isList,
				})
			}

			data := map[string]any{
				"model":  modelPath.Model,
				"fields": fields,
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, data, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	return cmd
}

func clientModelEndpoint(server, org, project, db, model string) string {
	return fmt.Sprintf("%s/graphql/end-user/org/%s/project/%s/db/%s/model/%s", strings.TrimRight(server, "/"), org, project, db, model)
}

func flattenGraphQLType(t struct {
	Kind   string `json:"kind"`
	Name   *string `json:"name"`
	OfType *struct {
		Kind   string `json:"kind"`
		Name   *string `json:"name"`
		OfType *struct {
			Kind string  `json:"kind"`
			Name *string `json:"name"`
		} `json:"ofType"`
	} `json:"ofType"`
}) (kind, typeName string, required, isList bool) {
	kind = t.Kind
	if t.Name != nil {
		typeName = *t.Name
	}

	if t.Kind == "NON_NULL" {
		required = true
		if t.OfType != nil {
			kind = t.OfType.Kind
			if t.OfType.Name != nil {
				typeName = *t.OfType.Name
			}
			if t.OfType.Kind == "LIST" {
				isList = true
				if t.OfType.OfType != nil && t.OfType.OfType.Name != nil {
					typeName = *t.OfType.OfType.Name
				}
			}
		}
		return kind, typeName, required, isList
	}

	if t.Kind == "LIST" {
		isList = true
		if t.OfType != nil && t.OfType.Name != nil {
			typeName = *t.OfType.Name
		}
	}

	return kind, typeName, required, isList
}
