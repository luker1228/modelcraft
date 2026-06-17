package rls

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	domainrls "modelcraft/internal/domain/rls"
)

type PolicyExpressionSQLCompiler struct{}

func NewPolicyExpressionSQLCompiler() *PolicyExpressionSQLCompiler {
	return &PolicyExpressionSQLCompiler{}
}

func (c *PolicyExpressionSQLCompiler) CompileUsing(
	_ context.Context,
	expr string,
	userCtx *domainrls.UserContext,
) (*domainrls.CompiledPolicy, error) {
	if strings.Contains(expr, "input.") {
		return nil, fmt.Errorf("input is not allowed in using expression")
	}

	sql, params, err := compileSimpleCELWhere(expr, userCtx)
	if err != nil {
		return nil, err
	}
	return &domainrls.CompiledPolicy{SQL: sql, Params: params}, nil
}

func compileSimpleCELWhere(expr string, userCtx *domainrls.UserContext) (string, []interface{}, error) {
	parser := newCELWhereParser(expr, userCtx)
	return parser.parse()
}

type celTokenKind string

const (
	celTokenIdentifier celTokenKind = "identifier"
	celTokenString     celTokenKind = "string"
	celTokenNumber     celTokenKind = "number"
	celTokenBool       celTokenKind = "bool"
	celTokenNull       celTokenKind = "null"
	celTokenOperator   celTokenKind = "operator"
	celTokenLParen     celTokenKind = "("
	celTokenRParen     celTokenKind = ")"
	celTokenLBracket   celTokenKind = "["
	celTokenRBracket   celTokenKind = "]"
	celTokenComma      celTokenKind = ","
	celTokenEOF        celTokenKind = "eof"
)

type celToken struct {
	kind  celTokenKind
	value string
}

type celWhereParser struct {
	tokens  []celToken
	pos     int
	userCtx *domainrls.UserContext
}

type celOperandKind string

const (
	celOperandField      celOperandKind = "field"
	celOperandValue      celOperandKind = "value"
	celOperandList       celOperandKind = "list"
	celOperandMethodCall celOperandKind = "method_call"
)

type celOperand struct {
	kind       celOperandKind
	field      string
	methodName string // only set for celOperandMethodCall
	value      interface{}
	list       []interface{}
}

func newCELWhereParser(expr string, userCtx *domainrls.UserContext) *celWhereParser {
	return &celWhereParser{tokens: tokenizeCELWhere(expr), userCtx: userCtx}
}

func (p *celWhereParser) parse() (string, []interface{}, error) {
	sql, params, err := p.parseOr()
	if err != nil {
		return "", nil, err
	}
	if p.peek().kind != celTokenEOF {
		return "", nil, fmt.Errorf("unexpected token %q", p.peek().value)
	}
	return sql, params, nil
}

func (p *celWhereParser) parseOr() (string, []interface{}, error) {
	leftSQL, leftParams, err := p.parseAnd()
	if err != nil {
		return "", nil, err
	}

	for p.peek().kind == celTokenOperator && p.peek().value == "||" {
		p.next()
		rightSQL, rightParams, err := p.parseAnd()
		if err != nil {
			return "", nil, err
		}
		leftSQL = "(" + leftSQL + " OR " + rightSQL + ")"
		leftParams = append(leftParams, rightParams...)
	}

	return leftSQL, leftParams, nil
}

func (p *celWhereParser) parseAnd() (string, []interface{}, error) {
	leftSQL, leftParams, err := p.parseUnary()
	if err != nil {
		return "", nil, err
	}

	for p.peek().kind == celTokenOperator && p.peek().value == "&&" {
		p.next()
		rightSQL, rightParams, err := p.parseUnary()
		if err != nil {
			return "", nil, err
		}
		leftSQL = "(" + leftSQL + " AND " + rightSQL + ")"
		leftParams = append(leftParams, rightParams...)
	}

	return leftSQL, leftParams, nil
}

func (p *celWhereParser) parseUnary() (string, []interface{}, error) {
	if p.peek().kind == celTokenOperator && p.peek().value == "!" {
		p.next()
		sql, params, err := p.parseUnary()
		if err != nil {
			return "", nil, err
		}
		return "NOT " + wrapSQL(sql), params, nil
	}

	if p.peek().kind == celTokenLParen {
		p.next()
		sql, params, err := p.parseOr()
		if err != nil {
			return "", nil, err
		}
		if _, err := p.expect(celTokenRParen, ")"); err != nil {
			return "", nil, err
		}
		return wrapSQL(sql), params, nil
	}

	return p.parseComparison()
}

func (p *celWhereParser) parseComparison() (string, []interface{}, error) {
	left, err := p.parseOperand()
	if err != nil {
		return "", nil, err
	}

	// Method calls (e.g. row.col.startsWith("x")) are self-contained boolean expressions.
	if left.kind == celOperandMethodCall {
		return buildMethodCallSQL(left)
	}

	op := p.peek()
	if op.kind != celTokenOperator {
		return "", nil, fmt.Errorf("expected comparison operator, got %q", op.value)
	}
	if !isComparisonOperator(op.value) {
		return "", nil, fmt.Errorf("unsupported operator %q", op.value)
	}
	p.next()

	right, err := p.parseOperand()
	if err != nil {
		return "", nil, err
	}

	return buildComparisonSQL(left, op.value, right)
}

func (p *celWhereParser) parseOperand() (celOperand, error) {
	token := p.next()
	switch token.kind {
	case celTokenIdentifier:
		switch {
		case strings.HasPrefix(token.value, "row."):
			rest := strings.TrimPrefix(token.value, "row.")
			if rest == "" {
				return celOperand{}, fmt.Errorf("unsupported row reference %q", token.value)
			}
			// row.field.method(args) — method call form
			if dotIdx := strings.LastIndex(rest, "."); dotIdx >= 0 {
				field, method := rest[:dotIdx], rest[dotIdx+1:]
				if field == "" || strings.Contains(field, ".") || method == "" {
					return celOperand{}, fmt.Errorf("unsupported row reference %q", token.value)
				}
				args, err := p.parseCallArgs()
				if err != nil {
					return celOperand{}, err
				}
				return celOperand{kind: celOperandMethodCall, field: field, methodName: method, list: args}, nil
			}
			// plain row.field
			return celOperand{kind: celOperandField, field: rest}, nil
		case strings.HasPrefix(token.value, "auth."):
			ref := strings.TrimPrefix(token.value, "auth.")
			if ref == "" || strings.Contains(ref, ".") {
				return celOperand{}, fmt.Errorf("unsupported auth reference %q", token.value)
			}
			return celOperand{kind: celOperandValue, value: resolveAuthValue(p.userCtx, ref)}, nil
		case strings.HasPrefix(token.value, "input."):
			return celOperand{}, fmt.Errorf("input is not allowed in using expression")
		default:
			return celOperand{}, fmt.Errorf("unsupported identifier %q", token.value)
		}
	case celTokenString:
		return celOperand{kind: celOperandValue, value: token.value}, nil
	case celTokenNumber:
		n, err := strconv.ParseFloat(token.value, 64)
		if err != nil {
			return celOperand{}, err
		}
		if strings.Contains(token.value, ".") {
			return celOperand{kind: celOperandValue, value: n}, nil
		}
		return celOperand{kind: celOperandValue, value: int64(n)}, nil
	case celTokenBool:
		return celOperand{kind: celOperandValue, value: token.value == "true"}, nil
	case celTokenNull:
		return celOperand{kind: celOperandValue, value: nil}, nil
	case celTokenLBracket:
		var values []interface{}
		if p.peek().kind == celTokenRBracket {
			p.next()
			return celOperand{kind: celOperandList, list: values}, nil
		}
		for {
			item, err := p.parseOperand()
			if err != nil {
				return celOperand{}, err
			}
			if item.kind != celOperandValue {
				return celOperand{}, fmt.Errorf("array literal supports only scalar values")
			}
			values = append(values, item.value)
			if p.peek().kind == celTokenComma {
				p.next()
				continue
			}
			if _, err := p.expect(celTokenRBracket, "]"); err != nil {
				return celOperand{}, err
			}
			break
		}
		return celOperand{kind: celOperandList, list: values}, nil
	default:
		return celOperand{}, fmt.Errorf("unexpected token %q", token.value)
	}
}

func buildComparisonSQL(left celOperand, op string, right celOperand) (string, []interface{}, error) {
	if left.kind != celOperandField {
		return "", nil, fmt.Errorf("left side must be a row field")
	}

	switch op {
	case "in":
		if right.kind != celOperandList {
			return "", nil, fmt.Errorf("IN requires an array literal")
		}
		if len(right.list) == 0 {
			return "1=0", nil, nil
		}
		placeholders := make([]string, 0, len(right.list))
		params := make([]interface{}, 0, len(right.list))
		for _, item := range right.list {
			placeholders = append(placeholders, "?")
			params = append(params, item)
		}
		return fmt.Sprintf("%s IN (%s)", left.field, strings.Join(placeholders, ", ")), params, nil
	case "==":
		if right.kind != celOperandValue {
			return "", nil, fmt.Errorf("right side must be a scalar value")
		}
		if right.value == nil {
			return left.field + " IS NULL", nil, nil
		}
		return left.field + " = ?", []interface{}{right.value}, nil
	case "!=":
		if right.kind != celOperandValue {
			return "", nil, fmt.Errorf("right side must be a scalar value")
		}
		if right.value == nil {
			return left.field + " IS NOT NULL", nil, nil
		}
		return left.field + " <> ?", []interface{}{right.value}, nil
	case ">", ">=", "<", "<=":
		if right.kind != celOperandValue {
			return "", nil, fmt.Errorf("right side must be a scalar value")
		}
		return left.field + " " + op + " ?", []interface{}{right.value}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator %q", op)
	}
}

// parseCallArgs consumes "(arg1, arg2, ...)" and returns the scalar values.
func (p *celWhereParser) parseCallArgs() ([]interface{}, error) {
	if _, err := p.expect(celTokenLParen, "("); err != nil {
		return nil, fmt.Errorf("method call expects '(': %w", err)
	}
	var args []interface{}
	if p.peek().kind == celTokenRParen {
		p.next()
		return args, nil
	}
	for {
		item, err := p.parseOperand()
		if err != nil {
			return nil, err
		}
		if item.kind != celOperandValue {
			return nil, fmt.Errorf("method call arguments must be scalar values")
		}
		args = append(args, item.value)
		if p.peek().kind == celTokenComma {
			p.next()
			continue
		}
		if _, err := p.expect(celTokenRParen, ")"); err != nil {
			return nil, err
		}
		break
	}
	return args, nil
}

// buildMethodCallSQL translates CEL string-method calls to SQL LIKE / REGEXP.
//
//	row.col.startsWith("x")  →  col LIKE 'x%'
//	row.col.endsWith("x")    →  col LIKE '%x'
//	row.col.contains("x")    →  col LIKE '%x%'
//	row.col.matches("x")     →  col REGEXP ?   (MySQL)
func buildMethodCallSQL(op celOperand) (string, []interface{}, error) {
	if len(op.list) != 1 {
		return "", nil, fmt.Errorf("method %q requires exactly one argument", op.methodName)
	}
	arg, ok := op.list[0].(string)
	if !ok {
		return "", nil, fmt.Errorf("method %q argument must be a string", op.methodName)
	}
	col := op.field
	switch op.methodName {
	case "startsWith":
		return col + " LIKE ?", []interface{}{arg + "%"}, nil
	case "endsWith":
		return col + " LIKE ?", []interface{}{"%" + arg}, nil
	case "contains":
		return col + " LIKE ?", []interface{}{"%" + arg + "%"}, nil
	case "matches":
		return col + " REGEXP ?", []interface{}{arg}, nil
	default:
		return "", nil, fmt.Errorf("unsupported string method %q", op.methodName)
	}
}

func tokenizeCELWhere(expr string) []celToken {
	tokens := make([]celToken, 0, len(expr))
	for i := 0; i < len(expr); {
		switch ch := expr[i]; {
		case ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r':
			i++
		case ch == '(':
			tokens = append(tokens, celToken{kind: celTokenLParen, value: "("})
			i++
		case ch == ')':
			tokens = append(tokens, celToken{kind: celTokenRParen, value: ")"})
			i++
		case ch == '[':
			tokens = append(tokens, celToken{kind: celTokenLBracket, value: "["})
			i++
		case ch == ']':
			tokens = append(tokens, celToken{kind: celTokenRBracket, value: "]"})
			i++
		case ch == ',':
			tokens = append(tokens, celToken{kind: celTokenComma, value: ","})
			i++
		case ch == '"' || ch == '\'':
			start := i
			i++
			for i < len(expr) {
				if expr[i] == '\\' {
					i += 2
					continue
				}
				if expr[i] == ch {
					break
				}
				i++
			}
			if i >= len(expr) {
				tokens = append(tokens, celToken{kind: celTokenString, value: expr[start+1:]})
				break
			}
			raw := expr[start : i+1]
			value, err := strconv.Unquote(raw)
			if err != nil {
				value = expr[start+1 : i]
			}
			tokens = append(tokens, celToken{kind: celTokenString, value: value})
			i++
		case i+1 < len(expr) && expr[i:i+2] == "&&":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "&&"})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "||":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "||"})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "==":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "=="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "!=":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "!="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == ">=":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: ">="})
			i += 2
		case i+1 < len(expr) && expr[i:i+2] == "<=":
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "<="})
			i += 2
		case ch == '!':
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "!"})
			i++
		case ch == '>':
			tokens = append(tokens, celToken{kind: celTokenOperator, value: ">"})
			i++
		case ch == '<':
			tokens = append(tokens, celToken{kind: celTokenOperator, value: "<"})
			i++
		case isIdentifierStart(ch):
			start := i
			i++
			for i < len(expr) && isIdentifierPart(expr[i]) {
				i++
			}
			value := expr[start:i]
			switch value {
			case "true", "false":
				tokens = append(tokens, celToken{kind: celTokenBool, value: value})
			case "null":
				tokens = append(tokens, celToken{kind: celTokenNull, value: value})
			case "in":
				tokens = append(tokens, celToken{kind: celTokenOperator, value: value})
			default:
				tokens = append(tokens, celToken{kind: celTokenIdentifier, value: value})
			}
		case isDigit(ch):
			start := i
			i++
			for i < len(expr) && (isDigit(expr[i]) || expr[i] == '.') {
				i++
			}
			tokens = append(tokens, celToken{kind: celTokenNumber, value: expr[start:i]})
		default:
			tokens = append(tokens, celToken{kind: celTokenOperator, value: string(ch)})
			i++
		}
	}
	tokens = append(tokens, celToken{kind: celTokenEOF, value: ""})
	return tokens
}

func isComparisonOperator(op string) bool {
	switch op {
	case "==", "!=", ">", ">=", "<", "<=", "in":
		return true
	default:
		return false
	}
}

func resolveAuthValue(userCtx *domainrls.UserContext, ref string) interface{} {
	if userCtx == nil {
		return ""
	}
	switch ref {
	case "roles":
		return userCtx.Roles
	case "userid", "uid", "user_id":
		// Preserve the native type (int64 or string) so SQL parameter binding is type-safe.
		return userCtx.UserIDValue()
	default:
		return userCtx.ResolveVariable(ref)
	}
}

func wrapSQL(sql string) string {
	if strings.HasPrefix(sql, "(") && strings.HasSuffix(sql, ")") {
		return sql
	}
	return "(" + sql + ")"
}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isIdentifierPart(ch byte) bool {
	return isIdentifierStart(ch) || isDigit(ch) || ch == '.'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (p *celWhereParser) peek() celToken {
	if p.pos >= len(p.tokens) {
		return celToken{kind: celTokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *celWhereParser) next() celToken {
	token := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return token
}

func (p *celWhereParser) expect(kind celTokenKind, value string) (celToken, error) {
	token := p.next()
	if token.kind != kind || (value != "" && token.value != value) {
		return celToken{}, fmt.Errorf("expected %q, got %q", value, token.value)
	}
	return token, nil
}
