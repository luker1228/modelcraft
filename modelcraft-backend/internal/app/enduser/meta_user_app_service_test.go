package enduser_test

import (
	"context"
	"modelcraft/internal/app/enduser"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 所有测试只校验参数验证逻辑，不触达 DB，因此传 nil *sql.DB 是安全的。

func TestMetaUserFindMany_TakeExceedsLimit_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    51,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "take must be <= 50")
}

func TestMetaUserFindMany_SkipExceedsLimit_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		Skip:    1001,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be <= 1000")
}

func TestMetaUserFindMany_NegativeSkip_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Skip:    -1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be >= 0")
}

func TestMetaUserFindMany_UnsupportedOrderByField_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		OrderBy: []enduser.MetaUserOrderByField{{CreatedAt: nil}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported orderBy field")
}

func TestMetaUserFindMany_InvalidSortDirection_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	invalidDir := "random"
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		OrderBy: []enduser.MetaUserOrderByField{{CreatedAt: &invalidDir}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort direction")
}

func TestMetaUserFindMany_MissingOrgName_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "",
		Take:    10,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName is required")
}

func TestMetaUserFindOne_NoCondition_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindOne(context.Background(), enduser.MetaUserFindOneCommand{
		OrgName: "test-org",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id or username is required")
}

func TestMetaUserFindOne_MissingOrgName_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindOne(context.Background(), enduser.MetaUserFindOneCommand{
		OrgName: "",
		ID:      "some-id",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName is required")
}

func TestMetaUserFindMany_ZeroTake_NoValidationError(t *testing.T) {
	// take 为 0 时默认为 20，不应触发 "take must be" 校验错误。
	// 校验通过后会走到 DB 调用（nil DB 会 panic），用 recover 隔离。
	svc := enduser.NewMetaUserAppService(nil)
	var err error
	func() {
		defer func() { recover() }() //nolint:errcheck // intentional: only testing validation path
		_, err = svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
			OrgName: "test-org",
			Take:    0,
		})
	}()
	if err != nil {
		assert.NotContains(t, err.Error(), "take must be")
	}
}
