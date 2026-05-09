package sqlsoftdelete

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const (
	deletedAtExpr   = "CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)"
	deleteTokenExpr = "CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)"
)

func RewriteFile(policy *Policy, src []byte) ([]byte, bool, error) {
	if policy == nil {
		return nil, false, fmt.Errorf("policy is required")
	}

	blocks := SplitSQLCFile(string(src))
	if len(blocks) == 0 {
		return src, false, nil
	}

	fileAnns := ParseAnnotations(src)
	leading := extractLeadingComments(string(src))

	var out bytes.Buffer
	if leading != "" {
		out.WriteString(strings.TrimRight(leading, "\n"))
		out.WriteString("\n\n")
	}

	changed := false
	for i, block := range blocks {
		anns := mergeAnnotations(fileAnns, block.Annotations)
		rewritten, blockChanged, err := RewriteBlock(policy, anns, block)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", block.Header, err)
		}
		if i > 0 {
			out.WriteString("\n\n")
		}
		out.WriteString(block.Header)
		out.WriteString("\n")
		out.WriteString(strings.TrimSpace(rewritten))
		changed = changed || blockChanged
	}

	return bytes.TrimSpace(out.Bytes()), changed, nil
}

func RewriteBlock(policy *Policy, anns Annotations, block QueryBlock) (string, bool, error) {
	if policy == nil {
		return "", false, fmt.Errorf("policy is required")
	}
	original := strings.TrimSpace(block.Body)

	parsed, err := ParseSQLBlock(block.Body)
	if err != nil {
		return "", false, err
	}

	if anns.IncludeDeleted || anns.OnlyDeleted || anns.PhysicalDelete {
		return original, false, nil
	}
	// sqlc.slice(...) is expanded by sqlc's internal text-edit pipeline. For now
	// keep these statements unchanged to avoid generator edit-boundary failures.
	if strings.Contains(strings.ToLower(original), "sqlc.slice(") {
		return original, false, nil
	}

	// Keep original SQL text as rewrite base so sqlc-specific syntax such as
	// sqlc.slice(...) is preserved byte-for-byte unless we explicitly patch it.
	rewritten := original
	switch parsed.Kind {
	case StatementKindSelect:
		for _, table := range parsed.Tables {
			if !policy.SoftDeleteEnabled(table.Name) {
				continue
			}
			currParsed, err := ParseSQLBlock(rewritten)
			if err != nil {
				return "", false, err
			}
			if currParsed.HasDeletedAtPredicate(table.AliasOrName()) {
				continue
			}
			qualifier := quoteIdent(table.AliasOrName())
			pred := qualifier + ".`deleted_at` = 0"
			next, ok := injectPredicateIntoSelect(rewritten, pred)
			if !ok {
				return "", false, fmt.Errorf("cannot inject predicate for table %s", table.Name)
			}
			rewritten = next
		}
	case StatementKindDelete:
		if len(parsed.Tables) == 0 {
			return "", false, fmt.Errorf("delete has no table")
		}
		main := parsed.Tables[0]
		if !policy.SoftDeleteEnabled(main.Name) {
			break
		}
		next, ok := rewriteDeleteSQL(rewritten, main, policy.NeedsDeleteToken(main.Name))
		if !ok {
			return "", false, fmt.Errorf("cannot rewrite delete statement")
		}
		rewritten = next
	default:
		return original, false, nil
	}

	rewritten = strings.TrimSpace(rewritten)
	return rewritten, rewritten != original, nil
}

func extractLeadingComments(src string) string {
	lines := strings.Split(src, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "-- name:") {
			break
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			out = append(out, line)
			continue
		}
		out = nil
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

func mergeAnnotations(a, b Annotations) Annotations {
	return Annotations{
		IncludeDeleted: a.IncludeDeleted || b.IncludeDeleted,
		OnlyDeleted:    a.OnlyDeleted || b.OnlyDeleted,
		PhysicalDelete: a.PhysicalDelete || b.PhysicalDelete,
	}
}

func quoteIdent(name string) string {
	trimmed := strings.TrimSpace(strings.Trim(name, "`"))
	if trimmed == "" {
		return "``"
	}
	return "`" + trimmed + "`"
}

func injectPredicateIntoSelect(sql, predicate string) (string, bool) {
	orderingRe := regexp.MustCompile(`(?is)\b(order\s+by|limit|for\s+update|lock\s+in\s+share\s+mode)\b`)
	whereRe := regexp.MustCompile(`(?is)\bwhere\b`)
	insertAt := len(sql)
	if loc := orderingRe.FindStringIndex(sql); loc != nil {
		insertAt = loc[0]
	}

	prefix := strings.TrimRight(sql[:insertAt], " \t\r\n")
	suffix := sql[insertAt:]
	// Keep terminal semicolon at tail, never in the predicate insertion segment.
	if strings.HasSuffix(prefix, ";") {
		prefix = strings.TrimSuffix(prefix, ";")
		if suffix == "" {
			suffix = ";"
		} else {
			suffix = ";" + suffix
		}
	}

	if whereLoc := whereRe.FindStringIndex(prefix); whereLoc != nil {
		if suffix != "" && !strings.HasPrefix(suffix, " ") && !strings.HasPrefix(suffix, "\n") && !strings.HasPrefix(suffix, "\t") {
			suffix = " " + suffix
		}
		return prefix + " AND " + predicate + suffix, true
	}
	if suffix != "" && !strings.HasPrefix(suffix, " ") && !strings.HasPrefix(suffix, "\n") && !strings.HasPrefix(suffix, "\t") {
		suffix = " " + suffix
	}
	return prefix + " WHERE " + predicate + suffix, true
}

var deleteStmtPattern = regexp.MustCompile(`(?is)^DELETE\s+FROM\s+(.+?)(?:\s+WHERE\s+(.+))?$`)

func rewriteDeleteSQL(sql string, table TableRef, includeDeleteToken bool) (string, bool) {
	sql = stripLeadingLineComments(sql)
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "UPDATE ") {
		return sql, false
	}

	trimmed := strings.TrimSpace(sql)
	if semi := strings.Index(trimmed, ";"); semi >= 0 {
		trimmed = strings.TrimSpace(trimmed[:semi])
	}

	matches := deleteStmtPattern.FindStringSubmatch(trimmed)
	if len(matches) == 0 {
		return "", false
	}

	tableExpr := strings.TrimSpace(matches[1])
	whereExpr := ""
	if len(matches) > 2 {
		whereExpr = strings.TrimSpace(matches[2])
		whereExpr = strings.TrimSuffix(whereExpr, ";")
		whereExpr = strings.TrimSpace(whereExpr)
	}

	qualifier := quoteIdent(table.AliasOrName())
	activePred := qualifier + ".`deleted_at` = 0"

	setClauses := []string{"`deleted_at` = " + deletedAtExpr}
	if includeDeleteToken {
		setClauses = append(setClauses, "`delete_token` = "+deleteTokenExpr)
	}

	var where string
	if whereExpr == "" {
		where = activePred
	} else if strings.Contains(strings.ToLower(whereExpr), "deleted_at") {
		where = whereExpr
	} else {
		where = "(" + whereExpr + ") AND " + activePred
	}

	rewritten := "UPDATE " + tableExpr + " SET " + strings.Join(setClauses, ", ") + " WHERE " + where + ";"
	return rewritten, true
}

func stripLeadingLineComments(sql string) string {
	lines := strings.Split(sql, "\n")
	start := 0
	for start < len(lines) {
		trimmed := strings.TrimSpace(lines[start])
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			start++
			continue
		}
		break
	}
	return strings.Join(lines[start:], "\n")
}
