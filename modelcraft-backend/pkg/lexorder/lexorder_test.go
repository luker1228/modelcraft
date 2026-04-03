package lexorder_test

import (
	"modelcraft/pkg/lexorder"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMidpoint(t *testing.T) {
	tests := []struct {
		name   string
		prev   string
		next   string
		wantGT string // result must be > wantGT
		wantLT string // result must be < wantLT
	}{
		{
			name:   "between two values",
			prev:   "a0",
			next:   "a2",
			wantGT: "a0",
			wantLT: "a2",
		},
		{
			name:   "between adjacent values",
			prev:   "a0",
			next:   "a1",
			wantGT: "a0",
			wantLT: "a1",
		},
		{
			name:   "after last (next empty = tail)",
			prev:   "a1",
			next:   "",
			wantGT: "a1",
			wantLT: "z",
		},
		{
			name:   "before first (prev empty = head)",
			prev:   "",
			next:   "a1",
			wantGT: "",
			wantLT: "a1",
		},
		{
			name:   "both empty returns initial value",
			prev:   "",
			next:   "",
			wantGT: "",
			wantLT: "z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := lexorder.Midpoint(tt.prev, tt.next)
			require.NoError(t, err)
			assert.Greater(t, result, tt.wantGT, "result should be > %q", tt.wantGT)
			assert.Less(t, result, tt.wantLT, "result should be < %q", tt.wantLT)
		})
	}
}

func TestMidpoint_Idempotent(t *testing.T) {
	// Multiple insertions between the same two values should always produce valid midpoints
	prev := "a0"
	next := "a1"
	for i := 0; i < 20; i++ {
		mid, err := lexorder.Midpoint(prev, next)
		require.NoError(t, err)
		assert.Greater(t, mid, prev)
		assert.Less(t, mid, next)
		next = mid
	}
}

func TestInitialOrder(t *testing.T) {
	order := lexorder.InitialOrder()
	assert.NotEmpty(t, order)
}

func TestRenumber(t *testing.T) {
	orders, err := lexorder.Renumber(5)
	require.NoError(t, err)
	assert.Len(t, orders, 5)

	// verify they are strictly ascending
	for i := 1; i < len(orders); i++ {
		assert.Greater(t, orders[i], orders[i-1], "orders[%d] should be > orders[%d]", i, i-1)
	}
}

func TestRenumber_Zero(t *testing.T) {
	orders, err := lexorder.Renumber(0)
	require.NoError(t, err)
	assert.Empty(t, orders)
}
