package projectgraphql

import (
	"context"
	"errors"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"

	appEnduser "modelcraft/internal/app/enduser"
	domainProject "modelcraft/internal/domain/project"
)

// FindUsers is the resolver for the findUsers field.
func (r *queryResolver) FindUsers(
	ctx context.Context, where *generated.UserWhereInput, skip, take *int32,
) (*generated.UserFindManyResult, error) {
	reqID := ctxutils.GetRequestID(ctx)

	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil || orgName == "" {
		return nil, newGQLError("organization context required", "MISSING_ORGANIZATION")
	}

	projectSlug, _ := ctxutils.GetProjectSlugFromContext(ctx)

	cmd := appEnduser.MetaUserFindManyCommand{
		ProjectScope: domainProject.ProjectScope{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
		},
	}
	if skip != nil {
		cmd.Skip = int(*skip)
	}
	if take != nil {
		cmd.Take = int(*take)
	}
	if where != nil {
		cmd.Where = convertUserWhereInput(where)
	}

	result, err := r.MetaUserAppService.FindMany(ctx, cmd)
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			return nil, newGQLError(bizErr.Msg(), bizErr.Info().GetCode())
		}
		return nil, err
	}

	items := make([]*generated.User, 0, len(result.Items))
	for _, dto := range result.Items {
		items = append(items, &generated.User{
			ID:        dto.ID,
			Username:  dto.Username,
			CreatedAt: dto.CreatedAt,
		})
	}

	return &generated.UserFindManyResult{
		Items: items,
		ReqID: reqID,
	}, nil
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*generated.UserFindOneResult, error) {
	reqID := ctxutils.GetRequestID(ctx)

	if !ctxutils.IsEndUser(ctx) {
		return nil, newGQLError("me is only available for end-user callers", "INVALID_CALLER")
	}

	dto, err := r.MetaUserAppService.GetMe(ctx)
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			return nil, newGQLError(bizErr.Msg(), bizErr.Info().GetCode())
		}
		return nil, err
	}
	if dto == nil {
		return &generated.UserFindOneResult{ReqID: reqID}, nil
	}

	return &generated.UserFindOneResult{
		Item:  &generated.User{ID: dto.ID, Username: dto.Username, CreatedAt: dto.CreatedAt},
		ReqID: reqID,
	}, nil
}

// convertUserWhereInput maps GraphQL UserWhereInput to app-layer filter.
func convertUserWhereInput(w *generated.UserWhereInput) *appEnduser.MetaUserFindManyFilter {
	if w == nil {
		return nil
	}
	f := &appEnduser.MetaUserFindManyFilter{}

	if w.ID != nil {
		f.IDEq = w.ID.Eq
		if len(w.ID.In) > 0 {
			f.IDIn = w.ID.In
		}
	}
	if w.Username != nil {
		f.UsernameEq = w.Username.Eq
		f.UsernameContains = w.Username.Contains
		f.UsernameStartsWith = w.Username.StartsWith
		if len(w.Username.In) > 0 {
			f.UsernameIn = w.Username.In
		}
	}
	if w.CreatedAt != nil {
		f.CreatedAtEq = w.CreatedAt.Eq
		f.CreatedAtGte = w.CreatedAt.Gte
		f.CreatedAtLte = w.CreatedAt.Lte
	}
	return f
}
