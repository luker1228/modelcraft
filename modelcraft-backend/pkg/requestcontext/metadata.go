package requestcontext

import (
	"context"
	"modelcraft/pkg/bizutils"
	"time"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const metadataContextKey contextKey = "request_metadata"

// RequestMetadata contains request-scoped metadata for tracking and observability
type RequestMetadata struct {
	ReqID     string    // Request tracking ID (UUID v7)
	StartTime time.Time // Request start timestamp
}

// WithMetadata creates a new context with RequestMetadata injected
func WithMetadata(ctx context.Context) context.Context {
	reqID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		// Fallback to timestamp-based ID if UUID generation fails
		reqID = time.Now().Format("20060102150405.000000")
	}

	metadata := &RequestMetadata{
		ReqID:     reqID,
		StartTime: time.Now(),
	}

	return context.WithValue(ctx, metadataContextKey, metadata)
}

// GetMetadata retrieves RequestMetadata from context
// Returns nil if metadata is not present
func GetMetadata(ctx context.Context) *RequestMetadata {
	if metadata, ok := ctx.Value(metadataContextKey).(*RequestMetadata); ok {
		return metadata
	}
	return nil
}

// CalculateTimeCost calculates the time elapsed since request start in milliseconds
// Returns 0 if metadata is not present in context
func CalculateTimeCost(ctx context.Context) int {
	metadata := GetMetadata(ctx)
	if metadata == nil {
		return 0
	}
	return int(time.Since(metadata.StartTime).Milliseconds())
}
