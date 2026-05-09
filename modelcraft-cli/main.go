package main

import (
	"os"

	"modelcraft-cli/cmd"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	os.Exit(cmd.Execute(cmd.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}, os.Args[1:], os.Stdout, os.Stderr))
}
