// Package runtime provides RLS (Row Level Security) resolution for runtime queries.
package runtime

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/http/middleware"
	"modelcraft/pkg/logfacade"
)

// MatchingService 匹配引擎接口（app/rls.PolicyMatchingService 实现）
type MatchingService interface {
	ResolveUsing(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
	ValidateCheck(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, input map[string]any, userCtx *rls.UserContext) error
}

// RLSResolver resolves RLS policies using the multi-policy matching engine.
type RLSResolver struct {
	logger      logfacade.Logger
	matchingSvc MatchingService
}

// NewRLSResolver creates a new RLSResolver.
func NewRLSResolver(logger logfacade.Logger, matchingSvc MatchingService) *RLSResolver {
	return &RLSResolver{
		logger:      logger,
		matchingSvc: matchingSvc,
	}
}

// ResolveResult represents the result of RLS resolution.
type ResolveResult struct {
	UsingSQL    string
	UsingParams []interface{}
	CheckSQL    string
	CheckParams []interface{}
	ShouldApply bool
	DenyAll     bool
}

// Resolve resolves the RLS policy for the given action and model.
// Returns nil DenyAll=false ShouldApply=false for Developer access (no RLS).
// Returns DenyAll=true when no matching policy exists (default deny).
func (r *RLSResolver) Resolve(ctx context.Context, action rls.Action, modelID string) (*ResolveResult, error) {
	// Get end-user identity from context
	identity := middleware.GetEndUserIdentity(ctx)

	// No identity found - deny all to be safe
	if identity == nil {
		r.logger.Warn(ctx, "No end-user identity found in context")
		return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
	}

	// Developer JWT — no RLS applied
	if identity.IsDeveloper() {
		return &ResolveResult{ShouldApply: false}, nil
	}

	// EndUser JWT — apply RLS
	if identity.IsEndUser() {
		rctx, ok := getRuntimeContext(ctx)
		if !ok {
			r.logger.Warn(ctx, "No runtime context found")
			return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
		}

		// Get UserContext from headers
		userCtx := middleware.GetUserContext(ctx)
		if userCtx == nil {
			userCtx = &rls.UserContext{}
		}

		// Resolve USING expression
		usingSQL, usingParams, err := r.matchingSvc.ResolveUsing(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, action, userCtx)
		if err != nil {
			r.logger.Debug(ctx, "RLS policy denied", logfacade.Err(err))
			return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
		}

		return &ResolveResult{
			UsingSQL:    usingSQL,
			UsingParams: usingParams,
			ShouldApply: true,
		}, nil
	}

	// Unknown issuer - deny all
	r.logger.Warn(ctx, "Unknown JWT issuer", logfacade.String("issuer", identity.Issuer))
	return &ResolveResult{ShouldApply: true, DenyAll: true}, nil
}

func (r *RLSResolver) ValidateInput(ctx context.Context, modelID string, action modelruntime.Action, input map[string]any) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}
	rctx, ok := getRuntimeContext(ctx)
	if !ok {
		return fmt.Errorf("RLS CHECK violation: runtime context missing")
	}
	domainAction := rls.ActionCreate
	if action == modelruntime.ActionUpdate {
		domainAction = rls.ActionUpdate
	}
	return r.matchingSvc.ValidateCheck(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, domainAction, input, userCtx)
}

// ValidateInsert validates an insert operation against the RLS check expression.
func (r *RLSResolver) ValidateInsert(ctx context.Context, modelID string, input map[string]interface{}) error {
	return r.ValidateInput(ctx, modelID, modelruntime.ActionInsert, input)
}

// ValidateUpdate validates an update operation against the RLS check expression.
func (r *RLSResolver) ValidateUpdate(ctx context.Context, modelID string, input map[string]interface{}) error {
	return r.ValidateInput(ctx, modelID, modelruntime.ActionUpdate, input)
}

// getRuntimeContext retrieves the runtime context from context.
func getRuntimeContext(ctx context.Context) (*runtimeContext, bool) {
	rctx, ok := ctx.Value(runtimeContextKey{}).(*runtimeContext)
	return rctx, ok
}
