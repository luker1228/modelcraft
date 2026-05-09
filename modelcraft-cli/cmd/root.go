package cmd

import "github.com/spf13/cobra"

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

func NewRootCommand(info BuildInfo) *cobra.Command {
	root := &cobra.Command{Use: "mc", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(newVersionCommand(info))
	return root
}
