package cmd

import (
	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

func newVersionCommand(info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:  "version",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			type versionData struct {
				Version   string `json:"version"`
				Commit    string `json:"commit"`
				BuildTime string `json:"buildTime"`
			}

			return output.WriteSuccess(cmd.OutOrStdout(), "json", true, versionData{
				Version:   info.Version,
				Commit:    info.Commit,
				BuildTime: info.BuildTime,
			}, nil)
		},
	}
}
