package main

import "modelcraft-cli/cmd"

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
	_ = root.Execute()
}
