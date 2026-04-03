package requestcontext

import (
	"context"
	"testing"
	"time"
)

func TestWithMetadata(t *testing.T) {
	ctx := context.Background()
	ctxWithMetadata := WithMetadata(ctx)

	metadata := GetMetadata(ctxWithMetadata)
	if metadata == nil {
		t.Fatal("Expected metadata to be present in context")
	}

	if metadata.ReqID == "" {
		t.Error("Expected ReqID to be non-empty")
	}

	if metadata.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	// Verify ReqID looks like a UUID (36 characters with hyphens)
	if len(metadata.ReqID) < 32 {
		t.Errorf("Expected ReqID to be UUID format, got: %s", metadata.ReqID)
	}
}

func TestGetMetadata_NotPresent(t *testing.T) {
	ctx := context.Background()
	metadata := GetMetadata(ctx)

	if metadata != nil {
		t.Error("Expected metadata to be nil when not present in context")
	}
}

func TestGetMetadata_Present(t *testing.T) {
	ctx := WithMetadata(context.Background())
	metadata := GetMetadata(ctx)

	if metadata == nil {
		t.Fatal("Expected metadata to be present")
	}

	// Verify we get the same instance
	metadata2 := GetMetadata(ctx)
	if metadata != metadata2 {
		t.Error("Expected to get same metadata instance from context")
	}
}

func TestCalculateTimeCost(t *testing.T) {
	ctx := WithMetadata(context.Background())

	// Wait a small amount of time
	time.Sleep(10 * time.Millisecond)

	timeCost := CalculateTimeCost(ctx)
	if timeCost < 10 {
		t.Errorf("Expected timeCost >= 10ms, got: %d", timeCost)
	}

	if timeCost > 1000 {
		t.Errorf("Expected timeCost < 1000ms, got: %d", timeCost)
	}
}

func TestCalculateTimeCost_NoMetadata(t *testing.T) {
	ctx := context.Background()
	timeCost := CalculateTimeCost(ctx)

	if timeCost != 0 {
		t.Errorf("Expected timeCost to be 0 when no metadata, got: %d", timeCost)
	}
}

func TestReqID_Uniqueness(t *testing.T) {
	// Generate multiple request IDs and verify uniqueness
	ids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		ctx := WithMetadata(context.Background())
		metadata := GetMetadata(ctx)

		if ids[metadata.ReqID] {
			t.Errorf("Duplicate ReqID found: %s", metadata.ReqID)
		}
		ids[metadata.ReqID] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestStartTime_Accuracy(t *testing.T) {
	before := time.Now()
	ctx := WithMetadata(context.Background())
	after := time.Now()

	metadata := GetMetadata(ctx)
	if metadata == nil {
		t.Fatal("Expected metadata to be present")
	}

	if metadata.StartTime.Before(before) {
		t.Error("StartTime is before context creation")
	}

	if metadata.StartTime.After(after) {
		t.Error("StartTime is after context creation")
	}
}

func BenchmarkWithMetadata(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WithMetadata(ctx)
	}
}

func BenchmarkGetMetadata(b *testing.B) {
	ctx := WithMetadata(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetMetadata(ctx)
	}
}

func BenchmarkCalculateTimeCost(b *testing.B) {
	ctx := WithMetadata(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateTimeCost(ctx)
	}
}
