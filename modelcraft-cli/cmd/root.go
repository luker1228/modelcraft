package cmd

import (
	"io"
	"strings"

	"modelcraft-cli/internal/output"

	"github.com/spf13/cobra"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

func NewRootCommand(info BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:   "mc",
		Short: "ModelCraft end-user CLI",
		Long:  "ModelCraft CLI for authentication, runtime model discovery, introspection, and GraphQL execution.",
		Example: "  mc auth login --server https://gateway.example.com --org acme --username alice --password '***'\n" +
			"  mc catalog projects\n" +
			"  mc describe sales.crm.users\n" +
			"  mc run sales.crm.users '{ findMany(take: 5) { id name } }'",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCommand(info))
	root.AddCommand(newAuthCommand())
	root.AddCommand(newCatalogCommand())
	root.AddCommand(newQueryCommand())
	root.AddCommand(newRunCommand())
	root.AddCommand(newSchemaCommand())
	root.AddCommand(newDescribeCommand())
	return root
}

func Execute(info BuildInfo, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(info)
	return executeCommand(root, args, stdout, stderr)
}

func executeCommand(root *cobra.Command, args []string, stdout, stderr io.Writer) int {
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	commandPath := commandPathForArgs(root, args)

	if err := root.Execute(); err != nil {
		normalized := normalizeError(commandPath, err)
		if writeErr := output.WriteError(stdout, "json", true, normalized); writeErr != nil {
			return output.ExitCode(writeErr)
		}
		return output.ExitCode(normalized)
	}

	return 0
}

func commandPathForArgs(root *cobra.Command, args []string) string {
	resolved, _, err := root.Find(args)
	if err != nil || resolved == nil {
		return root.CommandPath()
	}

	return resolved.CommandPath()
}

func normalizeError(commandPath string, err error) error {
	if err == nil {
		return nil
	}

	if normalized, ok := normalizeArgumentError(commandPath, err); ok {
		return normalized
	}

	return err
}

func normalizeArgumentError(commandPath string, err error) (error, bool) {
	code, message, ok := classifyArgumentError(err.Error())
	if !ok {
		return nil, false
	}

	details := map[string]any{"parserError": err.Error()}
	return output.NewCLIError(
		code,
		message,
		true,
		"Run '"+commandPath+" --help' to inspect valid arguments and flags.",
		details,
	), true
}

func classifyArgumentError(msg string) (code, message string, ok bool) {
	switch {
	case strings.Contains(msg, "required flag(s)"):
		return "MISSING_REQUIRED_FLAG", "Missing required flag.", true
	case strings.Contains(msg, "unknown flag"),
		strings.Contains(msg, "invalid argument"),
		strings.Contains(msg, "invalid value"),
		strings.Contains(msg, "flag needs an argument"):
		return "INVALID_ARGUMENT", "Invalid flag or flag value.", true
	case strings.Contains(msg, "unknown command"),
		strings.Contains(msg, "accepts 0 arg(s), received"):
		return "INVALID_ARGUMENT", "Unexpected argument or subcommand.", true
	}

	return "", "", false
}
