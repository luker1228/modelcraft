package projectgraphql

import (
	"context"
	"errors"
	modeldesign "modelcraft/internal/domain/modeldesign"
	rbacdomain "modelcraft/internal/domain/rbac"
	"modelcraft/pkg/bizerrors"
)

func toBizErr(err error) *bizerrors.BusinessError {
	var be *bizerrors.BusinessError
	if errors.As(err, &be) {
		return be
	}
	return bizerrors.NewError(bizerrors.SystemError, err.Error())
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// fetchModelMapForBundle collects model IDs from a single bundle's items and
// performs one batch lookup, returning a map keyed by model ID.
func fetchModelMapForBundle(
	ctx context.Context,
	r *queryResolver,
	orgName, projectSlug string,
	bundle *rbacdomain.EndUserPermissionBundle,
) map[string]*modeldesign.DataModel {
	if bundle == nil || len(bundle.Items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(bundle.Items))
	for _, item := range bundle.Items {
		ids = append(ids, item.ModelID)
	}
	modelMap, _ := r.ModelDesignService.GetModelMetaByIDs(ctx, orgName, projectSlug, ids)
	return modelMap
}
