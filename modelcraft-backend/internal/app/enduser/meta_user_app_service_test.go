package enduser_test

import (
	"context"
	"modelcraft/internal/app/enduser"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 所有测试只校验参数验证逻辑，不触达 DB，因此传 nil *sql.DB 是安全的。

func TestMetaUserFindMany_FirstExceedsLimit_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		First:   51,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first must be <= 50")
}

func TestMetaUserFindMany_InvalidCursor_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		First:   10,
		After:   "not-valid-base64!!!",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor")
}

func TestMetaUserFindMany_MissingOrgName_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(nil)
	_, err := svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
		OrgName: "",
		First:   10,
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

func TestMetaUserFindMany_ZeroFirst_NoValidationError(t *testing.T) {
	// first 为 0 时默认为 20，不应触发 "first must be" 校验错误。
	// 校验通过后会走到 DB 调用（nil DB 会 panic），用 recover 隔离。
	svc := enduser.NewMetaUserAppService(nil)
	var err error
	func() {
		defer func() { recover() }() //nolint:errcheck // intentional: only testing validation path
		_, err = svc.FindMany(context.Background(), enduser.MetaUserFindManyCommand{
			OrgName: "test-org",
			First:   0,
		})
	}()
	if err != nil {
		assert.NotContains(t, err.Error(), "first must be")
	}
}
