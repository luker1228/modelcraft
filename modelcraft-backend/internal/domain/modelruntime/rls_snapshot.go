// Package modelruntime provides domain types for model runtime execution.
package modelruntime

import (
	"context"

	"github.com/google/cel-go/cel"
)

// CheckProgram wraps a precompiled CEL program for CHECK expression evaluation.
// A nil CheckProgram means no CHECK is required (developer or true expression).
type CheckProgram struct {
	program cel.Program
}

// NewCheckProgram creates a CheckProgram from a precompiled CEL program.
func NewCheckProgram(program cel.Program) *CheckProgram {
	return &CheckProgram{program: program}
}

// Eval evaluates the CHECK expression against the given input and auth context.
// Returns nil if the expression evaluates to true, error otherwise.
func (c *CheckProgram) Eval(input map[string]any, auth map[string]any) error {
	if c == nil || c.program == nil {
		return nil
	}
	out, _, err := c.program.Eval(map[string]any{
		"input": input,
		"auth":  auth,
	})
	if err != nil {
		return err
	}
	allowed, ok := out.Value().(bool)
	if !ok {
		return &CheckEvalError{msg: "CHECK expression returned non-boolean"}
	}
	if !allowed {
		return &CheckEvalError{msg: "CHECK expression evaluated to false"}
	}
	return nil
}

// CheckEvalError represents a CHECK evaluation failure.
type CheckEvalError struct {
	msg string
}

func (e *CheckEvalError) Error() string {
	return "RLS CHECK violation: " + e.msg
}

// RLSPolicySnapshot holds pre-resolved RLS policies for the current request.
// Built once at Execute() entry, consumed by RLSInterceptDB at SQL execution time.
// nil snapshot means RLS is not applicable (developer JWT).
type RLSPolicySnapshot struct {
	// USING filters — injected into WHERE clause for SELECT/UPDATE/DELETE.
	// nil means no filtering needed (developer or true expression).
	SelectUSING *RawSQLFilter
	UpdateUSING *RawSQLFilter
	DeleteUSING *RawSQLFilter

	// CHECK programs — evaluated against input data before INSERT/UPDATE.
	// nil means no validation needed.
	InsertCHECK *CheckProgram
	UpdateCHECK *CheckProgram

	// Auth holds the pre-built auth context map for CEL evaluation.
	Auth map[string]any

	// DenyAll is true when no matching policy exists — all operations are denied.
	DenyAll bool
}

type rlsSnapshotKey struct{}

// WithRLSSnapshot stores the RLS snapshot on the context.
func WithRLSSnapshot(ctx context.Context, snap *RLSPolicySnapshot) context.Context {
	return context.WithValue(ctx, rlsSnapshotKey{}, snap)
}

// GetRLSSnapshot retrieves the RLS snapshot from the context.
// Returns nil if no snapshot was stored (developer access).
func GetRLSSnapshot(ctx context.Context) *RLSPolicySnapshot {
	snap, _ := ctx.Value(rlsSnapshotKey{}).(*RLSPolicySnapshot)
	return snap
}
