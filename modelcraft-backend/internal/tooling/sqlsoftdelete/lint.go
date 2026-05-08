package sqlsoftdelete

import (
	"fmt"
	"sort"
	"strings"
)

type Finding struct {
	File    string
	Query   string
	Message string
}

func LintFile(policy *Policy, file string, src []byte) ([]Finding, error) {
	if policy == nil {
		return nil, fmt.Errorf("policy is required")
	}

	blocks := SplitSQLCFile(string(src))
	findings := make([]Finding, 0, 8)

	for _, block := range blocks {
		parsed, err := ParseSQLBlock(block.Body)
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", file, block.Header, err)
		}

		if block.Annotations.IncludeDeleted || block.Annotations.OnlyDeleted {
			continue
		}

		for _, table := range parsed.Tables {
			if !policy.SoftDeleteEnabled(table.Name) {
				continue
			}

			if parsed.IsDelete() && !block.Annotations.PhysicalDelete {
				findings = append(findings, Finding{
					File:    file,
					Query:   block.Header,
					Message: fmt.Sprintf("physical DELETE on soft-delete table %s", table.Name),
				})
				continue
			}

			if parsed.IsSelect() && !parsed.HasDeletedAtPredicate(table.AliasOrName()) {
				findings = append(findings, Finding{
					File:    file,
					Query:   block.Header,
					Message: fmt.Sprintf("missing deleted_at predicate for %s", table.Name),
				})
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Query != findings[j].Query {
			return findings[i].Query < findings[j].Query
		}
		return findings[i].Message < findings[j].Message
	})

	return findings, nil
}

func RenderFindings(findings []Finding) string {
	if len(findings) == 0 {
		return ""
	}

	ordered := append([]Finding(nil), findings...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].File != ordered[j].File {
			return ordered[i].File < ordered[j].File
		}
		if ordered[i].Query != ordered[j].Query {
			return ordered[i].Query < ordered[j].Query
		}
		return ordered[i].Message < ordered[j].Message
	})

	var b strings.Builder
	for _, f := range ordered {
		b.WriteString(f.File)
		b.WriteString(" ")
		b.WriteString(f.Query)
		b.WriteString(": ")
		b.WriteString(f.Message)
		b.WriteString("\n")
	}
	return b.String()
}
