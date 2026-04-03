package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PermissionVersionManager manages version numbers for user+org permission sets.
// Version numbers are used for cache invalidation - incrementing a version causes
// old cache entries to naturally become stale.
//
// Phase 3 Implementation: Permission Versioning
// - Version key format: "perm_version:{orgName}:{userId}"
// - Version keys NEVER expire (persistent)
// - First access initializes version to 1
// - Thread-safe concurrent access via Redis INCR (atomic operation)
type PermissionVersionManager struct {
	redis *redis.Client
}

// NewPermissionVersionManager creates a new permission version manager
func NewPermissionVersionManager(redisClient *redis.Client) *PermissionVersionManager {
	return &PermissionVersionManager{
		redis: redisClient,
	}
}

// GetVersion retrieves the current version number for a user+org permission set.
// Returns 1 if version doesn't exist yet (first access).
// Timeout: 100ms for Redis operation.
func (m *PermissionVersionManager) GetVersion(ctx context.Context, orgName, userID string) (int64, error) {
	versionKey := m.buildVersionKey(orgName, userID)

	// Add timeout for Redis operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Get current version from Redis
	version, err := m.redis.Get(timeoutCtx, versionKey).Int64()
	if err == redis.Nil {
		// Version doesn't exist yet - initialize to 1
		// Initialize version to 1 (best effort, no error if fails)
		go m.initializeVersion(versionKey)

		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get permission version: %w", err)
	}

	return version, nil
}

// IncrementVersion increments the version number for a user+org permission set.
// This invalidates all cached permissions for this user+org combination.
// Returns the new version number after increment.
// Timeout: 100ms for Redis operation.
func (m *PermissionVersionManager) IncrementVersion(ctx context.Context, orgName, userID string) (int64, error) {
	versionKey := m.buildVersionKey(orgName, userID)

	// Add timeout for Redis operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// INCR is atomic - thread-safe for concurrent increments
	newVersion, err := m.redis.Incr(timeoutCtx, versionKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment permission version: %w", err)
	}

	return newVersion, nil
}

// initializeVersion initializes a version key to 1 (best effort, async)
func (m *PermissionVersionManager) initializeVersion(versionKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use SetArgs with NX option (SET if Not eXists) to avoid overwriting existing version
	// Error is intentionally ignored - this is a best-effort async operation
	_, _ = m.redis.SetArgs(ctx, versionKey, 1, redis.SetArgs{Mode: "NX"}).Result()
}

// buildVersionKey generates the Redis key for permission version
// Format: "perm_version:{orgName}:{userId}"
func (m *PermissionVersionManager) buildVersionKey(orgName, userID string) string {
	return fmt.Sprintf("perm_version:%s:%s", orgName, userID)
}
