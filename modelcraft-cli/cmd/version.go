package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand(info BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "{\"ok\":true,\"data\":{\"version\":\"%s\",\"commit\":\"%s\",\"buildTime\":\"%s\"}}\n", info.Version, info.Commit, info.BuildTime)
			return err
		},
	}
}
