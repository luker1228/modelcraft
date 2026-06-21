package modelruntime

import (
	"context"
	"modelcraft/pkg/ctxutils"
	"testing"
)

func TestRequestIDFromContext_UsesHTTPContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxutils.HttpRequestContextKey, &ctxutils.HttpRequestContext{
		RequestId: "req-123",
	})

	if got := requestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("expected req-123, got %q", got)
	}
}
