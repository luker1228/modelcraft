package sqlsoftdelete

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultModeEnabled  = "enabled"
	DefaultModeDisabled = "disabled"
)

type AnnotationConfig struct {
	IncludeDeleted string `yaml:"include_deleted"`
	OnlyDeleted    string `yaml:"only_deleted"`
	PhysicalDelete string `yaml:"physical_delete"`
}

type Policy struct {
	DefaultMode       string           `yaml:"default_mode"`
	TimestampUnit     string           `yaml:"timestamp_unit"`
	Annotations       AnnotationConfig `yaml:"annotations"`
	LintPaths         []string         `yaml:"lint_paths"`
	BlacklistTables   []string         `yaml:"blacklist_tables"`
	DeleteTokenTables []string         `yaml:"delete_token_tables"`

	blacklistSet   map[string]struct{}
	deleteTokenSet map[string]struct{}
}

func LoadPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal policy yaml: %w", err)
	}

	p.DefaultMode = normalizeDefaultMode(p.DefaultMode)
	p.Annotations = normalizeAnnotations(p.Annotations)
	p.blacklistSet = toLowerSet(p.BlacklistTables)
	p.deleteTokenSet = toLowerSet(p.DeleteTokenTables)

	return &p, nil
}

func (p *Policy) SoftDeleteEnabled(table string) bool {
	if p == nil {
		return false
	}

	tableName := strings.ToLower(strings.TrimSpace(table))
	_, blacklisted := p.blacklistSet[tableName]

	switch normalizeDefaultMode(p.DefaultMode) {
	case DefaultModeDisabled:
		return false
	case DefaultModeEnabled:
		return !blacklisted
	default:
		return !blacklisted
	}
}

func (p *Policy) NeedsDeleteToken(table string) bool {
	if p == nil {
		return false
	}

	tableName := strings.ToLower(strings.TrimSpace(table))
	_, ok := p.deleteTokenSet[tableName]
	return ok
}

func normalizeDefaultMode(mode string) string {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "" {
		return DefaultModeEnabled
	}
	if m != DefaultModeEnabled && m != DefaultModeDisabled {
		return DefaultModeEnabled
	}
	return m
}

func normalizeAnnotations(in AnnotationConfig) AnnotationConfig {
	out := in
	if strings.TrimSpace(out.IncludeDeleted) == "" {
		out.IncludeDeleted = "@include_deleted"
	}
	if strings.TrimSpace(out.OnlyDeleted) == "" {
		out.OnlyDeleted = "@only_deleted"
	}
	if strings.TrimSpace(out.PhysicalDelete) == "" {
		out.PhysicalDelete = "@physical_delete"
	}
	return out
}

func toLowerSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		n := strings.ToLower(strings.TrimSpace(item))
		if n == "" {
			continue
		}
		out[n] = struct{}{}
	}
	return out
}
