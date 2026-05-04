package adapter

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBizCodeToResourceType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		code     string
		expected generated.ResourceType
	}{
		{"ModelNotFound", bizerrors.ModelNotFound.GetCode(), generated.ResourceTypeModel},
		{"ProjectNotFound", bizerrors.ProjectNotFound.GetCode(), generated.ResourceTypeProject},
		{"ClusterNotFound", bizerrors.ClusterNotFound.GetCode(), generated.ResourceTypeCluster},
		{"EnumNotFound", bizerrors.EnumNotFound.GetCode(), generated.ResourceTypeEnum},
		{"GroupNotFound", bizerrors.GroupNotFound.GetCode(), generated.ResourceTypeGroup},
		{"UserNotFound", bizerrors.UserNotFound.GetCode(), generated.ResourceTypeUser},
		{"ProfileNotFound", bizerrors.ProfileNotFound.GetCode(), generated.ResourceTypeProfile},
		{"OrganizationNotFound", bizerrors.OrganizationNotFound.GetCode(), generated.ResourceTypeOrganization},
		{"RoleNotFound", bizerrors.RoleNotFound.GetCode(), generated.ResourceTypeRole},
		{"EndUserNotFound", bizerrors.EndUserNotFound.GetCode(), generated.ResourceTypeEndUser},
		{
			"EndUserPermissionNotFound",
			bizerrors.EndUserPermissionNotFound.GetCode(),
			generated.ResourceTypeEndUserPermission,
		},
		{
			"EndUserPermissionBundleNotFound",
			bizerrors.EndUserPermissionBundleNotFound.GetCode(),
			generated.ResourceTypeEndUserPermissionBundle,
		},
		{
			"EndUserPermissionBundleSnapshotNotFound",
			bizerrors.EndUserPermissionBundleSnapshotNotFound.GetCode(),
			generated.ResourceTypeEndUserPermissionBundleSnapshot,
		},
		{"EndUserRoleNotFound", bizerrors.EndUserRoleNotFound.GetCode(), generated.ResourceTypeEndUserRole},
		{
			"EndUserNotFoundInProject",
			bizerrors.EndUserNotFoundInProject.GetCode(),
			generated.ResourceTypeEndUserInProject,
		},
		// 兜底：通用 NOT_FOUND（无子类型）
		{"GenericNotFound", bizerrors.NotFound.GetCode(), generated.ResourceTypeUnknown},
		// 兜底：未收录的细粒度码（FIELD、FK、MEMBERSHIP、RECORD 等）
		{"FieldNotFound", bizerrors.FieldNotFound.GetCode(), generated.ResourceTypeUnknown},
		{"FKNotFound", bizerrors.FKNotFound.GetCode(), generated.ResourceTypeUnknown},
		{"MembershipNotFound", bizerrors.MembershipNotFound.GetCode(), generated.ResourceTypeUnknown},
		{"RecordNotFound", bizerrors.RecordNotFound.GetCode(), generated.ResourceTypeUnknown},
		{"APIKeyNotFound", bizerrors.APIKeyNotFound.GetCode(), generated.ResourceTypeUnknown},
		// 兜底：完全未知的码
		{"UnknownCode", "NOT_FOUND.SOMETHING_NEW", generated.ResourceTypeUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := BizCodeToResourceType(tc.code)
			require.Equal(t, tc.expected, got,
				"BizCodeToResourceType(%q) = %q, want %q", tc.code, got, tc.expected)
		})
	}
}
