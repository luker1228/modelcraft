package sqlsoftdelete

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

type StatementKind string

const (
	StatementKindUnknown StatementKind = "unknown"
	StatementKindSelect  StatementKind = "select"
	StatementKindDelete  StatementKind = "delete"
)

type TableRef struct {
	Name  string
	Alias string
}

func (t TableRef) AliasOrName() string {
	if strings.TrimSpace(t.Alias) != "" {
		return t.Alias
	}
	return t.Name
}

type ParsedStatement struct {
	Kind   StatementKind
	SQL    string
	Tables []TableRef
}

func (p ParsedStatement) IsDelete() bool {
	return p.Kind == StatementKindDelete
}

func (p ParsedStatement) IsSelect() bool {
	return p.Kind == StatementKindSelect
}

func (p ParsedStatement) HasDeletedAtPredicate(aliasOrName string) bool {
	sql := strings.ToLower(p.SQL)
	if aliasOrName != "" {
		alias := strings.ToLower(strings.Trim(strings.TrimSpace(aliasOrName), "`"))
		if alias != "" {
			qualifiedForms := []string{
				"`" + alias + "`.`deleted_at`",
				"`" + alias + "`.deleted_at",
				alias + ".`deleted_at`",
				alias + ".deleted_at",
			}
			for _, form := range qualifiedForms {
				if strings.Contains(sql, form) {
					return true
				}
			}
		}
	}

	if strings.Contains(sql, ".`deleted_at`") || strings.Contains(sql, ".deleted_at") {
		// Fallback for complex AST aliases when the parser-produced alias differs
		// but query still contains at least one qualified deleted_at predicate.
		if len(p.Tables) > 1 {
			return true
		}
	}

	if len(p.Tables) == 1 {
		bareForms := []string{
			" deleted_at",
			" `deleted_at`",
			"(deleted_at",
			"(`deleted_at`",
			"deleted_at=",
			"`deleted_at`=",
		}
		for _, form := range bareForms {
			if strings.Contains(sql, form) {
				return true
			}
		}
		if strings.HasPrefix(sql, "deleted_at") || strings.HasPrefix(sql, "`deleted_at`") {
			return true
		}
	}

	return false
}

func ParseSQLBlock(sql string) (*ParsedStatement, error) {
	stmtSQL := strings.TrimSpace(sql)
	if stmtSQL == "" {
		return nil, fmt.Errorf("empty SQL block")
	}

	p := parser.New()
	stmts, _, err := p.Parse(stmtSQL, "", "")
	if err != nil {
		return nil, fmt.Errorf("parse sql block: %w", err)
	}
	if len(stmts) != 1 {
		return nil, fmt.Errorf("expected 1 statement, got %d", len(stmts))
	}

	node := stmts[0]
	rendered, err := restoreSQL(node)
	if err != nil {
		return nil, err
	}

	parsed := &ParsedStatement{Kind: StatementKindUnknown, SQL: rendered}
	switch s := node.(type) {
	case *ast.SelectStmt:
		parsed.Kind = StatementKindSelect
		if s.From != nil {
			parsed.Tables = collectTablesFromJoin(s.From.TableRefs)
		}
	case *ast.DeleteStmt:
		parsed.Kind = StatementKindDelete
		if s.TableRefs != nil {
			parsed.Tables = collectTablesFromJoin(s.TableRefs.TableRefs)
		}
	}

	return parsed, nil
}

func restoreSQL(node ast.Node) (string, error) {
	var buf bytes.Buffer
	ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &buf)
	if err := node.Restore(ctx); err != nil {
		return "", fmt.Errorf("restore AST to SQL: %w", err)
	}
	return buf.String(), nil
}

func collectTablesFromJoin(join *ast.Join) []TableRef {
	if join == nil {
		return nil
	}
	out := make([]TableRef, 0, 4)
	collectTablesFromResultSet(join, &out)
	return out
}

func collectTablesFromResultSet(node ast.ResultSetNode, out *[]TableRef) {
	switch n := node.(type) {
	case *ast.Join:
		collectTablesFromResultSet(n.Left, out)
		if n.Right != nil {
			collectTablesFromResultSet(n.Right, out)
		}
	case *ast.TableSource:
		switch src := n.Source.(type) {
		case *ast.TableName:
			alias := src.Name.String()
			if n.AsName.String() != "" {
				alias = n.AsName.String()
			}
			*out = append(*out, TableRef{Name: src.Name.String(), Alias: alias})
		case *ast.Join:
			collectTablesFromResultSet(src, out)
		}
	case *ast.TableName:
		*out = append(*out, TableRef{Name: n.Name.String(), Alias: n.Name.String()})
	}
}
