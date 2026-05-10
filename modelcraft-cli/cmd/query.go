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
	var credentialsPath, project, whereRaw, orderByRaw string
	var selectFields []string
	var take, skip int

	cmd := &cobra.Command{
		Use:   "query <path>",
		Short: "Query multiple records from a runtime model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			where, err := parseJSONFlag("where", whereRaw)
			if err != nil {
				return err
			}
			orderBy, err := parseJSONFlag("orderBy", orderByRaw)
			if err != nil {
				return err
			}

			result, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Query(
				cmd.Context(), creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model, creds.AccessToken,
				client.QueryOptions{Where: where, Select: selectFields, OrderBy: orderBy, Take: take, Skip: skip},
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"items": result.Items}, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	cmd.Flags().StringVar(&whereRaw, "where", "", "JSON where filter")
	cmd.Flags().StringSliceVar(&selectFields, "select", nil, "Selected fields")
	cmd.Flags().StringVar(&orderByRaw, "orderBy", "", "JSON orderBy expression")
	cmd.Flags().IntVar(&take, "take", 20, "Page size")
	cmd.Flags().IntVar(&skip, "skip", 0, "Records to skip")
	return cmd
}

func newGetCommand() *cobra.Command {
	var credentialsPath, project, whereRaw string
	var selectFields []string

	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Query a single record from a runtime model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			where, err := parseJSONFlag("where", whereRaw)
			if err != nil {
				return err
			}
			if len(where) == 0 {
				return output.NewCLIError("MISSING_REQUIRED_FLAG", "Missing required flag.", true, "Run 'mc get --help' to inspect required flags.", map[string]any{"flag": "where"})
			}

			item, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Get(
				cmd.Context(), creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model, creds.AccessToken,
				client.GetOptions{Where: where, Select: selectFields},
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, item, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	cmd.Flags().StringVar(&whereRaw, "where", "", "JSON where filter")
	cmd.Flags().StringSliceVar(&selectFields, "select", nil, "Selected fields")
	return cmd
}

func newCountCommand() *cobra.Command {
	var credentialsPath, project, whereRaw string

	cmd := &cobra.Command{
		Use:   "count <path>",
		Short: "Count records from a runtime model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			where, err := parseJSONFlag("where", whereRaw)
			if err != nil {
				return err
			}

			count, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Count(
				cmd.Context(), creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model, creds.AccessToken,
				client.CountOptions{Where: where},
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{"count": count}, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	cmd.Flags().StringVar(&whereRaw, "where", "", "JSON where filter")
	return cmd
}

func newAggregateCommand() *cobra.Command {
	var credentialsPath, project, whereRaw string
	var fields []string

	cmd := &cobra.Command{
		Use:   "aggregate <path>",
		Short: "Aggregate records from a runtime model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, modelPath, creds, err := resolveRuntimeContext(cmd, credentialsPath, project, args[0])
			if err != nil {
				return err
			}

			where, err := parseJSONFlag("where", whereRaw)
			if err != nil {
				return err
			}

			result, err := (client.GraphQLClient{HTTPClient: http.DefaultClient}).Aggregate(
				cmd.Context(), creds.Server, creds.OrgName, modelPath.Project, modelPath.Database, modelPath.Model, creds.AccessToken,
				client.AggregateOptions{Where: where, Fields: fields},
			)
			if err != nil {
				return err
			}
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, result, ctx)
		},
	}

	bindRuntimeSharedFlags(cmd, &credentialsPath, &project)
	cmd.Flags().StringVar(&whereRaw, "where", "", "JSON where filter")
	cmd.Flags().StringSliceVar(&fields, "fields", nil, "Aggregation fields")
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
