package dml

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// RLSInterceptDB wraps ClientDatabaseRepository to apply RLS policies
// before SQL execution. It reads RLSPolicySnapshot from context and
// transparently injects USING filters and evaluates CHECK expressions.
type RLSInterceptDB struct {
	inner modelruntime.ClientDatabaseRepository
}

// NewRLSInterceptDB creates a new RLS intercept wrapper.
func NewRLSInterceptDB(inner modelruntime.ClientDatabaseRepository) *RLSInterceptDB {
	return &RLSInterceptDB{inner: inner}
}

// injectUSING appends the USING filter to RawFilters if present.
func injectUSING(ctx context.Context, filters *[]modelruntime.RawSQLFilter, using *modelruntime.RawSQLFilter) {
	if using != nil {
		logfacade.GetLogger(ctx).Debugf(ctx, "RLS inject USING: sql=%s params=%d", using.SQL, len(using.Params))
		*filters = append(*filters, *using)
	}
}

// evalCHECKs evaluates CHECK expressions with OR logic:
// any single program passing is sufficient. Returns error only if all fail.
func evalCHECKs(
	ctx context.Context, checks []*modelruntime.CheckProgram, input, auth map[string]any,
) error {
	if len(checks) == 0 {
		return nil
	}
	logger := logfacade.GetLogger(ctx)
	for i, check := range checks {
		if check == nil {
			continue
		}
		if err := check.Eval(input, auth); err == nil {
			logger.Debugf(ctx, "RLS CHECK passed: index=%d/%d", i, len(checks))
			return nil // any one passing is sufficient
		}
	}
	logger.Debugf(ctx, "RLS CHECK failed: all %d checks rejected", len(checks))
	return bizerrors.NewError(bizerrors.PermissionDenied, "all CHECK expressions failed")
}

// errNoPolicy returns a permission denied error for the given action when no policy is configured.
func errNoPolicy(action string) error {
	return bizerrors.NewError(bizerrors.PermissionDenied, "no RLS policy configured for action: "+action)
}

// ---- Read operations ----

func (r *RLSInterceptDB) FindUnique(ctx context.Context, input *modelruntime.FindUniqueInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.FindUnique(ctx, input)
}

func (r *RLSInterceptDB) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.FindFirst(ctx, input)
}

func (r *RLSInterceptDB) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.FindMany(ctx, input)
}

func (r *RLSInterceptDB) ListByCursor(
	ctx context.Context, input *modelruntime.ListByCursorInput,
) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.ListByCursor(ctx, input)
}

func (r *RLSInterceptDB) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.Aggregate(ctx, input)
}

func (r *RLSInterceptDB) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoSelectPolicy {
			return nil, errNoPolicy("read")
		}
		injectUSING(ctx, &input.RawFilters, snap.SelectFilter)
	}
	return r.inner.Count(ctx, input)
}

// FindManyIn is used for N+1 relation loading, no RLS interception currently.
func (r *RLSInterceptDB) FindManyIn(
	ctx context.Context, input *modelruntime.FindManyInInput,
) ([]map[string]any, error) {
	return r.inner.FindManyIn(ctx, input)
}

// ---- INSERT operations ----

func (r *RLSInterceptDB) CreateOne(ctx context.Context, input *modelruntime.CreateOneInput) (string, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoCreatePolicy {
			return "", errNoPolicy("create")
		}
		if err := evalCHECKs(ctx, snap.CreateChecks, input.Data, snap.Auth); err != nil {
			return "", err
		}
	}
	return r.inner.CreateOne(ctx, input)
}

func (r *RLSInterceptDB) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoCreatePolicy {
			return nil, errNoPolicy("create")
		}
		for _, dataItem := range input.Data {
			if err := evalCHECKs(ctx, snap.CreateChecks, dataItem, snap.Auth); err != nil {
				return nil, err
			}
		}
	}
	return r.inner.CreateMany(ctx, input)
}

// ---- UPDATE operations ----

func (r *RLSInterceptDB) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoUpdatePolicy {
			return nil, errNoPolicy("update")
		}
		injectUSING(ctx, &input.RawFilters, snap.UpdateFilter)
		if err := evalCHECKs(ctx, snap.UpdateChecks, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateOne(ctx, input)
}

func (r *RLSInterceptDB) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoUpdatePolicy {
			return nil, errNoPolicy("update")
		}
		injectUSING(ctx, &input.RawFilters, snap.UpdateFilter)
		if err := evalCHECKs(ctx, snap.UpdateChecks, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateMany(ctx, input)
}

// ---- DELETE operations ----

func (r *RLSInterceptDB) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoDeletePolicy {
			return nil, errNoPolicy("delete")
		}
		injectUSING(ctx, &input.RawFilters, snap.DeleteFilter)
	}
	return r.inner.DeleteOne(ctx, input)
}

func (r *RLSInterceptDB) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if snap.NoDeletePolicy {
			return nil, errNoPolicy("delete")
		}
		injectUSING(ctx, &input.RawFilters, snap.DeleteFilter)
	}
	return r.inner.DeleteMany(ctx, input)
}
