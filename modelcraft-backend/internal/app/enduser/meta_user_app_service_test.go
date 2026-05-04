package enduser_test

import (
	"context"
	"database/sql"
	"modelcraft/internal/app/enduser"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDBManager 实现 enduser.PrivateDBManager 接口，
// 用于单元测试中绕过真实 DB 连接，测试纯校验逻辑。
type stubDBManager struct{}

func (s *stubDBManager) GetOrInit(_ context.Context, _, _ string) (*sql.DB, error) {
	return nil, nil //nolint:nilnil // stub: tests only exercise validation, not DB access
}

// --- FindMany 分页边界测试 ---

func TestMetaUserFindMany_TakeExceedsLimit_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    51, // 超过 50 上限
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "take must be <= 50")
}

func TestMetaUserFindMany_SkipExceedsLimit_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		Skip:    1001, // 超过 1000 上限
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be <= 1000")
}

func TestMetaUserFindMany_NegativeSkip_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Skip:    -1,
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "skip must be >= 0")
}

func TestMetaUserFindMany_UnsupportedOrderByField_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	// CreatedAt 为 nil 触发非白名单字段校验
	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		OrderBy: []enduser.MetaUserOrderByField{
			{CreatedAt: nil},
		},
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported orderBy field")
}

func TestMetaUserFindMany_InvalidSortDirection_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	invalidDir := "random"
	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    10,
		OrderBy: []enduser.MetaUserOrderByField{
			{CreatedAt: &invalidDir},
		},
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort direction")
}

func TestMetaUserFindMany_MissingOrgName_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "",
		Take:    10,
	}

	_, err := svc.FindMany(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName is required")
}

// --- FindOne 参数校验测试 ---

func TestMetaUserFindOne_NoCondition_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindOneCommand{
		OrgName: "test-org",
		// ID 和 Username 均为空
	}

	_, err := svc.FindOne(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id or username is required")
}

func TestMetaUserFindOne_MissingOrgName_ReturnsError(t *testing.T) {
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindOneCommand{
		OrgName: "",
		ID:      "some-id",
	}

	_, err := svc.FindOne(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName is required")
}

// --- Take 默认值测试（校验层不报错）---

func TestMetaUserFindMany_ZeroTake_NoValidationError(t *testing.T) {
	// take 为 0 时默认为 20，不应触发校验错误
	svc := enduser.NewMetaUserAppService(&stubDBManager{})
	ctx := context.Background()

	cmd := enduser.MetaUserFindManyCommand{
		OrgName: "test-org",
		Take:    0,
	}

	// stubDBManager 返回 nil *sql.DB，FindMany 在校验通过后会因 nil DB 而 panic。
	// 用 recover 确认校验阶段不产生 "take must be" 错误。
	var err error
	func() {
		defer func() { recover() }() //nolint:errcheck // intentional: only testing validation, not DB path
		_, err = svc.FindMany(ctx, cmd)
	}()
	if err != nil {
		assert.NotContains(t, err.Error(), "take must be")
	}
}
