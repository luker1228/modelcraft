package sqlsoftdelete

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

const defaultConfigPath = "db/soft_delete.yaml"

// Run executes the sqlsoftdelete command with subcommands.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "lint":
		return runLint(args[1:], stdout, stderr)
	case "codemod":
		return runCodemod(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown subcommand: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runLint(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", defaultConfigPath, "Path to soft delete policy file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	policy, err := LoadPolicy(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "load policy: %v\n", err)
		return 1
	}

	files, err := expandLintFiles(policy.LintPaths)
	if err != nil {
		fmt.Fprintf(stderr, "resolve lint paths: %v\n", err)
		return 1
	}
	if len(files) == 0 {
		fmt.Fprintln(stdout, "sqlsoftdelete lint: no SQL files matched")
		return 0
	}

	findings, err := lintFiles(policy, files)
	if err != nil {
		fmt.Fprintf(stderr, "lint failed: %v\n", err)
		return 1
	}
	if len(findings) > 0 {
		fmt.Fprint(stderr, RenderFindings(findings))
		fmt.Fprintf(stderr, "sqlsoftdelete lint failed: %d finding(s)\n", len(findings))
		return 1
	}

	fmt.Fprintf(stdout, "sqlsoftdelete lint passed: %d file(s) checked\n", len(files))
	return 0
}

func runCodemod(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("codemod", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", defaultConfigPath, "Path to soft delete policy file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	policy, err := LoadPolicy(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "load policy: %v\n", err)
		return 1
	}

	files, err := expandLintFiles(policy.LintPaths)
	if err != nil {
		fmt.Fprintf(stderr, "resolve codemod paths: %v\n", err)
		return 1
	}
	if len(files) == 0 {
		fmt.Fprintln(stdout, "sqlsoftdelete codemod: no SQL files matched")
		return 0
	}

	changed := 0
	for _, file := range files {
		src, readErr := os.ReadFile(file)
		if readErr != nil {
			fmt.Fprintf(stderr, "read %s: %v\n", file, readErr)
			return 1
		}

		rewritten, fileChanged, rewriteErr := RewriteFile(policy, src)
		if rewriteErr != nil {
			fmt.Fprintf(stderr, "rewrite %s: %v\n", file, rewriteErr)
			return 1
		}
		if !fileChanged {
			continue
		}

		content := append(rewritten, '\n')
		if writeErr := os.WriteFile(file, content, 0o644); writeErr != nil {
			fmt.Fprintf(stderr, "write %s: %v\n", file, writeErr)
			return 1
		}
		changed++
	}

	fmt.Fprintf(stdout, "sqlsoftdelete codemod finished: %d/%d file(s) changed\n", changed, len(files))
	return 0
}

func lintFiles(policy *Policy, files []string) ([]Finding, error) {
	all := make([]Finding, 0, 16)
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		findings, err := LintFile(policy, file, src)
		if err != nil {
			return nil, err
		}
		all = append(all, findings...)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		if all[i].Query != all[j].Query {
			return all[i].Query < all[j].Query
		}
		return all[i].Message < all[j].Message
	})
	return all, nil
}

func expandLintFiles(patterns []string) ([]string, error) {
	set := make(map[string]struct{}, 64)
	for _, pattern := range patterns {
		p := strings.TrimSpace(filepath.ToSlash(pattern))
		if p == "" {
			continue
		}

		matches, err := doublestar.FilepathGlob(p)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", p, err)
		}

		for _, match := range matches {
			info, statErr := os.Stat(match)
			if statErr != nil {
				if errors.Is(statErr, fs.ErrNotExist) {
					continue
				}
				return nil, fmt.Errorf("stat %q: %w", match, statErr)
			}
			if info.IsDir() {
				continue
			}
			set[match] = struct{}{}
		}
	}

	files := make([]string, 0, len(set))
	for file := range set {
		files = append(files, file)
	}
	sort.Strings(files)
	return files, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: sqlsoftdelete <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  lint     Check SQL queries for soft-delete rules")
	fmt.Fprintln(w, "  codemod  Rewrite SQL queries to soft-delete style")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Flags:\n  --config string\tPath to policy file (default %q)\n", defaultConfigPath)
}

// Main is the process entrypoint used by cmd/sqlsoftdelete/main.go.
func Main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
