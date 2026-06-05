package resource

import (
	"strings"

	"modelcraft-cli/internal/output"
)

type ParseContext struct {
	CurrentProject string
}

type ModelPath struct {
	Project  string
	Database string
	Model    string
}

func ParseModelPath(raw string, ctx ParseContext) (ModelPath, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	// Trim spaces from each segment so "luke .db.model" is not silently accepted.
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	switch len(parts) {
	case 3:
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			break
		}
		return ModelPath{Project: parts[0], Database: parts[1], Model: parts[2]}, nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			break
		}
		if ctx.CurrentProject == "" {
			return ModelPath{}, output.NewCLIError("NO_PROJECT_CONTEXT", "No project context is selected.", true, "Use --project <slug> or run 'mc auth switch-project <slug>'.", nil)
		}
		return ModelPath{Project: ctx.CurrentProject, Database: parts[0], Model: parts[1]}, nil
	}

	return ModelPath{}, output.NewCLIError(
		"INVALID_RESOURCE_PATH",
		"Resource path must be '<project>.<database>.<model>' or '<database>.<model>'.",
		true,
		"Provide at least database and model segments.",
		map[string]any{"path": raw},
	)
}
