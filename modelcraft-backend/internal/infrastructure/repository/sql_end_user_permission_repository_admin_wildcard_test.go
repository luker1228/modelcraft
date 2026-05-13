package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/infrastructure/dbgen"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type adminWildcardQuerierStub struct {
	dbgen.Querier

	roleIDs      []string
	roleByID     map[string]dbgen.EndUserRole
	explicitIDs  []string
	directIDs    []string
	explicitCall bool
	directCall   bool
}

func (s *adminWildcardQuerierStub) ListRolesByUser(
	_ context.Context,
	_ dbgen.ListRolesByUserParams,
) ([]string, error) {
	return s.roleIDs, nil
}

func (s *adminWildcardQuerierStub) GetEndUserRoleByID(
	_ context.Context,
	arg dbgen.GetEndUserRoleByIDParams,
) (dbgen.EndUserRole, error) {
	if role, ok := s.roleByID[arg.ID]; ok {
		return role, nil
	}
	return dbgen.EndUserRole{}, sql.ErrNoRows
}

func (s *adminWildcardQuerierStub) GetBundleIDsByUserExplicitRoles(
	_ context.Context,
	_ dbgen.GetBundleIDsByUserExplicitRolesParams,
) ([]string, error) {
	s.explicitCall = true
	return s.explicitIDs, nil
}

func (s *adminWildcardQuerierStub) GetBundleIDsByUserDirect(
	_ context.Context,
	_ dbgen.GetBundleIDsByUserDirectParams,
) ([]string, error) {
	s.directCall = true
	return s.directIDs, nil
}

func TestFindPermissionsByEndUserAndModel_AdminRoleReturnsWildcard(t *testing.T) {
	stub := &adminWildcardQuerierStub{
		roleIDs: []string{"role-admin"},
		roleByID: map[string]dbgen.EndUserRole{
			"role-admin": {
				ID:          "role-admin",
				OrgName:     "org-1",
				ProjectSlug: "proj-1",
				Name:        "admin",
				IsProtected: true,
			},
		},
	}
	repo := &SqlEndUserDataPermissionRepository{q: stub}

	perms, err := repo.FindPermissionsByEndUserAndModel(context.Background(), "org-1", "proj-1", "user-1", "model-1")
	require.NoError(t, err)
	require.Len(t, perms, 1)

	p := perms[0]
	require.NotNil(t, p.RowPolicy)
	assert.Equal(t, "model-1", p.ModelID)
	assert.True(t, p.RowPolicy.Select.Allowed)
	assert.True(t, p.RowPolicy.Insert.Allowed)
	assert.True(t, p.RowPolicy.Update.Allowed)
	assert.True(t, p.RowPolicy.Delete.Allowed)
	assert.Equal(t, "all", string(p.RowPolicy.Insert.Scope))

	assert.False(t, stub.explicitCall)
	assert.False(t, stub.directCall)
}

func TestFindPermissionsByEndUserAndModel_NonAdminFallsBackToBundleLookup(t *testing.T) {
	stub := &adminWildcardQuerierStub{
		roleIDs: []string{"role-editor"},
		roleByID: map[string]dbgen.EndUserRole{
			"role-editor": {
				ID:          "role-editor",
				OrgName:     "org-1",
				ProjectSlug: "proj-1",
				Name:        "editor",
				IsProtected: true,
			},
		},
		explicitIDs: nil,
		directIDs:   nil,
	}
	repo := &SqlEndUserDataPermissionRepository{q: stub}

	perms, err := repo.FindPermissionsByEndUserAndModel(context.Background(), "org-1", "proj-1", "user-1", "model-1")
	require.NoError(t, err)
	assert.Nil(t, perms)
	assert.True(t, stub.explicitCall)
	assert.True(t, stub.directCall)
}
