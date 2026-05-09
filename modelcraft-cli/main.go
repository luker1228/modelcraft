package main

import (
	"fmt"
	"os"

	"modelcraft-cli/cmd"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	root := cmd.NewRootCommand(cmd.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
