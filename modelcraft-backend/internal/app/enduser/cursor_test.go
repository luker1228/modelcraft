package enduser

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cursor_test.go 用白盒测试覆盖 encodeCursor / decodeCursor 私有函数。

func TestEncodeCursor_DecodeCursor_RoundTrip(t *testing.T) {
	ts := time.Date(2024, 5, 20, 15, 0, 0, 0, time.UTC)
	id := "abc-uuid-123"

	cursor := encodeCursor(ts, id)
	require.NotEmpty(t, cursor)

	gotTime, gotID, err := decodeCursor(cursor)
	require.NoError(t, err)
	assert.Equal(t, id, gotID)
	// UnixMilli 精度，纳秒部分会被截断
	assert.Equal(t, ts.UnixMilli(), gotTime.UnixMilli())
}

func TestEncodeCursor_SameTimeDifferentID_ProducesDifferentCursors(t *testing.T) {
	ts := time.Date(2024, 5, 20, 15, 0, 0, 0, time.UTC)

	c1 := encodeCursor(ts, "id-1")
	c2 := encodeCursor(ts, "id-2")
	assert.NotEqual(t, c1, c2)
}

func TestDecodeCursor_InvalidBase64_ReturnsError(t *testing.T) {
	_, _, err := decodeCursor("not-valid-base64!!!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base64 decode failed")
}

func TestDecodeCursor_MissingPipe_ReturnsError(t *testing.T) {
	// "this has no pipe" 是合法 Base64，但内容不含 |
	_, _, err := decodeCursor(base64.StdEncoding.EncodeToString([]byte("this has no pipe")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor format")
}

func TestDecodeCursor_InvalidTimestamp_ReturnsError(t *testing.T) {
	// 手动构造 Base64("not-a-number|some-id")
	raw := base64.StdEncoding.EncodeToString([]byte("not-a-number|some-id"))
	_, _, err := decodeCursor(raw)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor timestamp")
}

func TestEncodeCursor_ZeroTime_StillDecodable(t *testing.T) {
	zeroTime := time.Time{}
	id := "zero-time-id"

	cursor := encodeCursor(zeroTime, id)
	gotTime, gotID, err := decodeCursor(cursor)
	require.NoError(t, err)
	assert.Equal(t, id, gotID)
	assert.Equal(t, zeroTime.UnixMilli(), gotTime.UnixMilli())
}
