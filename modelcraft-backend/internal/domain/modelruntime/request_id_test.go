package modelruntime

import (
	"modelcraft/pkg/ctxutils"
	"testing"
)

func TestRequestIDFromContext_UsesHTTPContext(t *testing.T) {
	ctx := ctxutils.SetRequestID(t.Context(), "req-123")

	if got := requestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("expected req-123, got %q", got)
	}
}
