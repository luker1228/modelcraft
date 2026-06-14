package dml

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/pkg/bizerrors"
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
func injectUSING(filters *[]modelruntime.RawSQLFilter, using *modelruntime.RawSQLFilter) {
	if using != nil {
		*filters = append(*filters, *using)
	}
}

// evalCHECK evaluates a CHECK expression if present.
func evalCHECK(
	check *modelruntime.CheckProgram, input, auth map[string]any,
) error {
	if check == nil {
		return nil
	}
	if err := check.Eval(input, auth); err != nil {
		return bizerrors.NewError(bizerrors.PermissionDenied, err.Error())
	}
	return nil
}

// ---- Read operations ----

func (r *RLSInterceptDB) FindUnique(ctx context.Context, input *modelruntime.FindUniqueInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindUnique(ctx, input)
}

func (r *RLSInterceptDB) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindFirst(ctx, input)
}

func (r *RLSInterceptDB) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindMany(ctx, input)
}

func (r *RLSInterceptDB) ListByCursor(
	ctx context.Context, input *modelruntime.ListByCursorInput,
) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.ListByCursor(ctx, input)
}

func (r *RLSInterceptDB) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.Aggregate(ctx, input)
}

func (r *RLSInterceptDB) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
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
		if err := evalCHECK(snap.InsertCHECK, input.Data, snap.Auth); err != nil {
			return "", err
		}
	}
	return r.inner.CreateOne(ctx, input)
}

func (r *RLSInterceptDB) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		for _, dataItem := range input.Data {
			if err := evalCHECK(snap.InsertCHECK, dataItem, snap.Auth); err != nil {
				return nil, err
			}
		}
	}
	return r.inner.CreateMany(ctx, input)
}

// ---- UPDATE operations ----

func (r *RLSInterceptDB) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.UpdateUSING)
		if err := evalCHECK(snap.UpdateCHECK, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateOne(ctx, input)
}

func (r *RLSInterceptDB) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.UpdateUSING)
		if err := evalCHECK(snap.UpdateCHECK, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateMany(ctx, input)
}

// ---- DELETE operations ----

func (r *RLSInterceptDB) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.DeleteUSING)
	}
	return r.inner.DeleteOne(ctx, input)
}

func (r *RLSInterceptDB) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.DeleteUSING)
	}
	return r.inner.DeleteMany(ctx, input)
}
