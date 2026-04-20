// Package runtime provides RLS (Row Level Security) resolution for runtime queries.
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/http/middleware"
	"modelcraft/pkg/logfacade"
)

// RLSFilter represents the resolved RLS filter for a query.
type RLSFilter struct {
	SelectPredicate JSONExpr `json:"selectPredicate"`
	InsertCheck     JSONExpr `json:"insertCheck"`
	UpdatePredicate JSONExpr `json:"updatePredicate"`
	UpdateCheck     JSONExpr `json:"updateCheck"`
	DeletePredicate JSONExpr `json:"deletePredicate"`
	FieldName       string   `json:"fieldName"` // Fixed as "owner"
	EndUserID       string   `json:"endUserId"`
}

// JSONExpr represents a JSON expression for RLS predicates.
type JSONExpr string

// IsTrue returns true if the expression is a JSON boolean true.
func (e JSONExpr) IsTrue() bool {
	var v interface{}
	if err := json.Unmarshal([]byte(e), &v); err != nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	// Check {"_const": true}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err == nil {
		if v, ok := obj["_const"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}

// IsFalse returns true if the expression is a JSON boolean false.
func (e JSONExpr) IsFalse() bool {
	var v interface{}
	if err := json.Unmarshal([]byte(e), &v); err != nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return !b
	}
	// Check {"_const": false}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err == nil {
		if v, ok := obj["_const"]; ok {
			if b, ok := v.(bool); ok {
				return !b
			}
		}
	}
	return false
}

// IsOwnerEqualsUser returns true if the expression matches {"owner":{"_eq":{"_auth":"uid"}}}
func (e JSONExpr) IsOwnerEqualsUser() bool {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(e), &obj); err != nil {
		return false
	}
	owner, ok := obj["owner"].(map[string]interface{})
	if !ok {
		return false
	}
	eq, ok := owner["_eq"].(map[string]interface{})
	if !ok {
		return false
	}
	auth, ok := eq["_auth"].(string)
	return ok && auth == "uid"
}

// IsDenyAll returns true if all predicates are false (DENY ALL policy).
func (f *RLSFilter) IsDenyAll() bool {
	return f.SelectPredicate.IsFalse() &&
		f.UpdatePredicate.IsFalse() &&
		f.DeletePredicate.IsFalse()
}

// ShouldInjectWhere returns true if WHERE clause injection is needed.
// Returns false when SELECT predicate is true (full access) or false (DENY ALL).
func (f *RLSFilter) ShouldInjectWhere() bool {
	return !f.SelectPredicate.IsTrue() && !f.SelectPredicate.IsFalse()
}

// RLSResolver resolves RLS filters for runtime queries.
type RLSResolver struct {
	logger     logfacade.Logger
	policyRepo modeldesign.ModelRLSPolicyRepository
}

// NewRLSResolver creates a new RLSResolver.
func NewRLSResolver(logger logfacade.Logger, policyRepo modeldesign.ModelRLSPolicyRepository) *RLSResolver {
	return &RLSResolver{
		logger:     logger,
		policyRepo: policyRepo,
	}
}

// ResolveResult represents the result of RLS resolution.
type ResolveResult struct {
	// Filter is the resolved RLS filter. Nil means no filtering (Developer access).
	Filter *RLSFilter
	// ShouldApply indicates whether RLS should be applied.
	ShouldApply bool
	// DenyAll indicates if all access should be denied.
	DenyAll bool
}

// Resolve resolves the RLS filter based on the context and model.
// Returns nil filter for Developer JWT access (no RLS).
// Returns DENY ALL filter for EndUser when no policy exists.
func (r *RLSResolver) Resolve(ctx context.Context, modelID string) (*ResolveResult, error) {
	logger := r.logger

	// Get end-user identity from context
	identity := middleware.GetEndUserIdentity(ctx)

	// No identity found - this shouldn't happen for Runtime endpoints
	// but we return deny-all to be safe
	if identity == nil {
		logger.Warn(ctx, "No end-user identity found in context")
		return &ResolveResult{
			Filter:      nil,
			ShouldApply: true,
			DenyAll:     true,
		}, nil
	}

	// Developer JWT should not reach Runtime endpoints (rejected by middleware)
	// But if it does, we return deny-all as a safety measure
	if identity.IsDeveloper() {
		logger.Warn(ctx, "Developer JWT detected in Runtime endpoint")
		return &ResolveResult{
			Filter:      nil,
			ShouldApply: true,
			DenyAll:     true,
		}, nil
	}

	// EndUser JWT - apply RLS
	if identity.IsEndUser() {
		// Fetch model policy from repository
		// Get orgName and projectSlug from context
		rctx, ok := getRuntimeContext(ctx)
		if !ok {
			logger.Warn(ctx, "No runtime context found")
			return &ResolveResult{
				Filter:      nil,
				ShouldApply: true,
				DenyAll:     true,
			}, nil
		}

		// Fetch policy from repository
		policy, err := r.policyRepo.GetByModelID(ctx, rctx.OrgName, rctx.ProjectSlug, modelID)
		if err != nil {
			logger.Error(ctx, "Failed to fetch RLS policy", logfacade.Err(err))
			// On error, deny all to be safe
			return &ResolveResult{
				Filter:      nil,
				ShouldApply: true,
				DenyAll:     true,
			}, nil
		}

		// No policy found - DENY ALL (default deny)
		if policy == nil {
			logger.Debug(ctx, "No RLS policy found for model, denying all access",
				logfacade.String("modelID", modelID))
			return &ResolveResult{
				Filter:      nil,
				ShouldApply: true,
				DenyAll:     true,
			}, nil
		}

		// Convert domain policy to runtime filter
		filter := &RLSFilter{
			SelectPredicate: JSONExpr(policy.SelectPredicate),
			InsertCheck:     JSONExpr(policy.InsertCheck),
			UpdatePredicate: JSONExpr(policy.UpdatePredicate),
			UpdateCheck:     JSONExpr(policy.UpdateCheck),
			DeletePredicate: JSONExpr(policy.DeletePredicate),
			FieldName:       "owner",
			EndUserID:       identity.EndUserID,
		}

		// Check if this is DENY ALL (all predicates are false)
		isDenyAll := filter.SelectPredicate.IsFalse() &&
			filter.UpdatePredicate.IsFalse() &&
			filter.DeletePredicate.IsFalse()

		return &ResolveResult{
			Filter:      filter,
			ShouldApply: true,
			DenyAll:     isDenyAll,
		}, nil
	}

	// Unknown issuer - deny all
	logger.Warn(ctx, "Unknown JWT issuer", logfacade.String("issuer", identity.Issuer))
	return &ResolveResult{
		Filter:      nil,
		ShouldApply: true,
		DenyAll:     true,
	}, nil
}

// CompileToWhereClause compiles the RLS filter to a SQL WHERE clause.
// Returns the WHERE clause and parameters.
func (r *RLSResolver) CompileToWhereClause(filter *RLSFilter) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	// Handle true constant - no filtering needed
	if filter.SelectPredicate.IsTrue() {
		return "", nil, nil
	}

	// Handle false constant - return a condition that's always false
	if filter.SelectPredicate.IsFalse() {
		return "1=0", nil, nil
	}

	// Handle owner equals current user
	if filter.SelectPredicate.IsOwnerEqualsUser() {
		return "owner = ?", []interface{}{filter.EndUserID}, nil
	}

	// TODO: Implement full JSON expression compilation
	// For now, return an error for unsupported expressions
	return "", nil, fmt.Errorf("unsupported RLS expression: %s", filter.SelectPredicate)
}

// ValidateInsert validates an insert operation against the RLS policy.
func (r *RLSResolver) ValidateInsert(filter *RLSFilter, data map[string]interface{}) error {
	if filter == nil {
		return nil
	}

	// Handle true constant - allow all
	if filter.InsertCheck.IsTrue() {
		return nil
	}

	// Handle false constant - deny all
	if filter.InsertCheck.IsFalse() {
		return fmt.Errorf("RLS CHECK violation: INSERT not allowed")
	}

	// Handle owner equals current user
	if filter.InsertCheck.IsOwnerEqualsUser() {
		owner, ok := data["owner"]
		if !ok || owner == nil {
			// No owner specified - will be auto-filled
			return nil
		}
		ownerStr, ok := owner.(string)
		if !ok || ownerStr != filter.EndUserID {
			return fmt.Errorf("RLS CHECK violation: owner must be the current user")
		}
		return nil
	}

	// TODO: Implement full JSON expression validation
	return fmt.Errorf("unsupported RLS expression for INSERT: %s", filter.InsertCheck)
}

// ValidateUpdate validates an update operation against the RLS policy.
func (r *RLSResolver) ValidateUpdate(filter *RLSFilter, data map[string]interface{}) error {
	if filter == nil {
		return nil
	}

	// Handle true constant - allow all
	if filter.UpdateCheck.IsTrue() {
		return nil
	}

	// Handle false constant - deny all
	if filter.UpdateCheck.IsFalse() {
		return fmt.Errorf("RLS CHECK violation: UPDATE not allowed")
	}

	// Handle owner equals current user
	if filter.UpdateCheck.IsOwnerEqualsUser() {
		// Check if trying to change owner to another user
		if newOwner, ok := data["owner"]; ok {
			newOwnerStr, ok := newOwner.(string)
			if ok && newOwnerStr != filter.EndUserID {
				return fmt.Errorf("RLS CHECK violation: cannot change owner to another user")
			}
		}
		return nil
	}

	// TODO: Implement full JSON expression validation
	return fmt.Errorf("unsupported RLS expression for UPDATE: %s", filter.UpdateCheck)
}

// getRuntimeContext retrieves the runtime context from context.
func getRuntimeContext(ctx context.Context) (*runtimeContext, bool) {
	rctx, ok := ctx.Value(runtimeContextKey{}).(*runtimeContext)
	return rctx, ok
}

// AutoFillOwner adds the owner field to data if not present.
// Returns true if the field was added.
func (r *RLSResolver) AutoFillOwner(filter *RLSFilter, data map[string]interface{}) bool {
	if filter == nil {
		return false
	}

	// Only auto-fill if insert check requires owner equals current user
	if !filter.InsertCheck.IsOwnerEqualsUser() && !filter.InsertCheck.IsTrue() {
		return false
	}

	if _, ok := data["owner"]; !ok {
		data["owner"] = filter.EndUserID
		return true
	}

	// Override if different
	if data["owner"] != filter.EndUserID {
		data["owner"] = filter.EndUserID
		return true
	}

	return false
}
