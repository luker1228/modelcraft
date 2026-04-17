package orggraphql

import (
	"context"
	"strconv"

	"modelcraft/pkg/bizerrors"
)

func (r *mutationResolver) validateRoleIDs(ctx context.Context, orgName string, roleIDs []string) ([]int, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	if r.RoleService == nil {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"role service not available",
		)
	}

	roles, err := r.RoleService.ListRoles(ctx, orgName, true)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int]struct{}, len(roles))
	for _, role := range roles {
		allowed[role.ID] = struct{}{}
	}

	parsed := make([]int, 0, len(roleIDs))
	seen := make(map[int]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		id, parseErr := strconv.Atoi(roleID)
		if parseErr != nil {
			return nil, bizerrors.NewErrorFromContext(
				ctx,
				bizerrors.ParamInvalid,
				"invalid role id: %s",
				roleID,
			)
		}
		if _, ok := allowed[id]; !ok {
			return nil, bizerrors.NewErrorFromContext(
				ctx,
				bizerrors.ParamInvalid,
				"role id not found in organization: %s",
				roleID,
			)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		parsed = append(parsed, id)
	}

	return parsed, nil
}
