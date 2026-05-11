package cmd

import (
	"modelcraft-cli/internal/output"
	"modelcraft-cli/internal/schema"

	"github.com/spf13/cobra"
)

func newSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "schema", Short: "Export local CLI schema"}
	cmd.AddCommand(newSchemaCommandsCommand())
	return cmd
}

func newSchemaCommandsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commands",
		Short: "Export command and flag schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc := schema.BuildCommandSchema()
			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, doc, nil)
		},
	}
}
