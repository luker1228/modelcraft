package endusergraphql

import (
	"context"

	appModelDesign "modelcraft/internal/app/modeldesign"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
)

// Empty implements the _empty query for schema health check.
func (r *queryResolver) Empty(ctx context.Context) (*string, error) {
	s := "ok"
	return &s, nil
}

// DatabaseCatalog implements the databaseCatalog query.
func (r *queryResolver) DatabaseCatalog(
	ctx context.Context,
	input *generated.DatabaseCatalogInput,
) (*generated.GetDatabaseCatalogPayload, error) {
	requestID := ctxutils.GetRequestID(ctx)

	// Extract context from context (set by middleware)
	orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
	projectSlug, _ := ctxutils.GetProjectSlugFromContext(ctx)

	if orgName == "" || projectSlug == "" {
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.InvalidInput{
				Message: "Organization and project context required",
			},
		}, nil
	}

	// Set defaults
	page := 1
	pageSize := 20
	search := ""

	if input != nil {
		if input.Page != nil && *input.Page > 0 {
			page = int(*input.Page)
		}
		if input.PageSize != nil && *input.PageSize > 0 {
			pageSize = int(*input.PageSize)
			if pageSize > 100 {
				pageSize = 100
			}
		}
		if input.Search != nil {
			search = *input.Search
		}
	}

	databases, totalCount, err := r.ModelDesignService.QueryDatabaseCatalogWithCommand(ctx, appModelDesign.DatabaseCatalogQueryCommand{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Search:      search,
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		return r.mapDatabaseCatalogError(ctx, requestID, err), nil
	}

	// Convert to GraphQL types
	items := make([]*generated.DatabaseLite, len(databases))
	for i, db := range databases {
		items[i] = &generated.DatabaseLite{Name: db}
	}

	page32 := int32(page)
	pageSize32 := int32(pageSize)
	totalCount32 := int32(totalCount)

	return &generated.GetDatabaseCatalogPayload{
		Data: &generated.DatabaseCatalogPayload{
			Databases:  items,
			TotalCount: totalCount32,
			Page:       page32,
			PageSize:   pageSize32,
		},
	}, nil
}

// ModelCatalog implements the modelCatalog query.
func (r *queryResolver) ModelCatalog(
	ctx context.Context,
	input generated.ModelCatalogInput,
) (*generated.GetModelCatalogPayload, error) {
	requestID := ctxutils.GetRequestID(ctx)

	// Extract context from context (set by middleware)
	orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
	projectSlug, _ := ctxutils.GetProjectSlugFromContext(ctx)

	if orgName == "" || projectSlug == "" {
		return &generated.GetModelCatalogPayload{
			Error: generated.InvalidInput{
				Message: "Organization and project context required",
			},
		}, nil
	}

	databaseName := input.DatabaseName
	if databaseName == "" {
		return &generated.GetModelCatalogPayload{
			Error: generated.InvalidInput{
				Message: "databaseName is required",
			},
		}, nil
	}

	// Set defaults
	page := 1
	pageSize := 50
	search := ""

	if input.Page != nil && *input.Page > 0 {
		page = int(*input.Page)
	}
	if input.PageSize != nil && *input.PageSize > 0 {
		pageSize = int(*input.PageSize)
		if pageSize > 100 {
			pageSize = 100
		}
	}
	if input.Search != nil {
		search = *input.Search
	}

	models, totalCount, err := r.ModelDesignService.QueryModelsWithCommand(ctx, appModelDesign.ModelQueryCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
		Name:         search,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		return r.mapModelCatalogError(ctx, requestID, err), nil
	}

	// Convert to GraphQL types
	items := make([]*generated.ModelLite, len(models))
	for i, m := range models {
		items[i] = &generated.ModelLite{
			ID:           m.ID,
			Name:         m.ModelName,
			Title:        m.Title,
			DatabaseName: m.DatabaseName,
		}
	}

	page32 := int32(page)
	pageSize32 := int32(pageSize)
	totalCount32 := int32(totalCount)

	return &generated.GetModelCatalogPayload{
		Data: &generated.ModelCatalogPayload{
			Models:     items,
			TotalCount: totalCount32,
			Page:       page32,
			PageSize:   pageSize32,
		},
	}, nil
}

func (r *queryResolver) mapDatabaseCatalogError(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetDatabaseCatalogPayload {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		// System error - wrap as internal
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	code := bizErr.Info().GetCode()
	switch code {
	case bizerrors.ProjectNotFound.Code:
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: bizErr.Msg(),
			},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.Unauthorized{
				Message: bizErr.Msg(),
			},
		}
	default:
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.InvalidInput{
				Message: bizErr.Msg(),
			},
		}
	}
}

func (r *queryResolver) mapModelCatalogError(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetModelCatalogPayload {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		// System error - wrap as internal
		return &generated.GetModelCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	code := bizErr.Info().GetCode()
	switch code {
	case bizerrors.ProjectNotFound.Code:
		return &generated.GetModelCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: bizErr.Msg(),
			},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetModelCatalogPayload{
			Error: generated.Unauthorized{
				Message: bizErr.Msg(),
			},
		}
	default:
		return &generated.GetModelCatalogPayload{
			Error: generated.InvalidInput{
				Message: bizErr.Msg(),
			},
		}
	}
}
