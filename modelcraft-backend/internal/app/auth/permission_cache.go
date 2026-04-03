package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/pkg/logfacade"
	"time"

	"github.com/redis/go-redis/v9"
)

// PermissionCacheInterface defines the interface for cached permission loading.
// This enables mocking in tests and provides a contract for cache implementations.
type PermissionCacheInterface interface {
	// GetUserPermissions loads only permissions (legacy method, kept for backward compatibility)
	GetUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

	// GetUserPermissionsAndRoles loads both roles and their permissions in map structure
	// Returns: map[roleName]*RolePermissionInfo where each role contains its permissions
	GetUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (RolePermissions, error)
}

// PermissionCache wraps PermissionLoader with Redis caching for performance optimization.
// Phase 2 Implementation: Redis caching with 5-minute TTL
// Phase 3 Implementation: Version-based cache invalidation
// Performance Target: <1ms for cache hit, <50ms for cache miss
type PermissionCache struct {
	redis          *redis.Client
	permLoader     PermissionLoaderInterface
	versionManager *PermissionVersionManager
	cacheTTL       time.Duration
}

// Ensure PermissionCache implements PermissionCacheInterface
var _ PermissionCacheInterface = (*PermissionCache)(nil)

// NewPermissionCache creates a new permission cache with Redis backing and version manager
func NewPermissionCache(
	redisClient *redis.Client,
	permLoader PermissionLoaderInterface,
	versionManager *PermissionVersionManager,
	cacheTTL time.Duration,
) *PermissionCache {
	return &PermissionCache{
		redis:          redisClient,
		permLoader:     permLoader,
		versionManager: versionManager,
		cacheTTL:       cacheTTL,
	}
}

// GetUserPermissions retrieves user permissions from cache or loads from database.
// Phase 3: Cache Key Format: "auth:{orgName}:{userId}:{version}"
// Cache TTL: 5 minutes (configurable)
//
// Flow:
// 1. Get current version number from VersionManager
// 2. Try to get from Redis cache using versioned key
// 3. If cache hit: deserialize and return (<1ms)
// 4. If cache miss (or version mismatch): load from DB via PermissionLoader
// 5. Store result in cache with current version number
// 6. Return permissions
//
// Version-based Invalidation:
// - When permissions change, version number is incremented
// - Old cache entries become stale automatically (version mismatch)
// - No manual cache clearing needed
//
// Error Handling:
// - Redis connection errors are logged but don't fail the request (falls back to DB)
// - Database errors are propagated to the caller
func (c *PermissionCache) GetUserPermissions(ctx context.Context, userID, orgName string) ([]string, error) {
	// Step 1: Get current version number
	version, err := c.versionManager.GetVersion(ctx, orgName, userID)
	if err != nil {
		// Log error but continue with version 1 (fallback)
		version = 1
	}

	cacheKey := c.buildVersionedCacheKey(orgName, userID, version)

	// Step 2: Try cache first (with timeout)
	cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cached, err := c.redis.Get(cacheCtx, cacheKey).Result()
	if err == nil {
		// Cache hit - deserialize permissions
		var permissions []string
		if jsonErr := json.Unmarshal([]byte(cached), &permissions); jsonErr != nil {
			// Continue to DB load on deserialization error
		} else {
			return permissions, nil
		}
	} else if err != redis.Nil {
		// Redis error (not a cache miss) - log and continue to DB
		logfacade.GetLogger(ctx).Warn(ctx, "Redis cache error, falling back to database", logfacade.Err(err))
	}

	// Step 3: Cache miss - load from database
	permissions, err := c.permLoader.LoadUserPermissions(ctx, userID, orgName)
	if err != nil {
		return nil, err
	}

	// Step 4: Store in cache with current version (best effort - don't fail if cache write fails)
	go c.cachePermissions(cacheKey, permissions)

	return permissions, nil
}

// GetUserPermissionsAndRoles retrieves user roles and permissions from cache or loads from database.
// Phase 3: Cache Key Format: "auth:{orgName}:{userId}:{version}"
// Cache Value Format: JSON map of {"roleName": {"permissions": [...]}}
// Cache TTL: 5 minutes (configurable)
//
// Flow:
// 1. Get current version number from VersionManager
// 2. Try to get from Redis cache using versioned key
// 3. If cache hit: deserialize map and return (<1ms)
// 4. If cache miss (or version mismatch): load from DB via PermissionLoader
// 5. Store result in cache with current version number
// 6. Return RolePermissions map
//
// Version-based Invalidation:
// - When permissions change, version number is incremented
// - Old cache entries become stale automatically (version mismatch)
// - No manual cache clearing needed
//
// Error Handling:
// - Redis connection errors are logged but don't fail the request (falls back to DB)
// - Database errors are propagated to the caller
func (c *PermissionCache) GetUserPermissionsAndRoles(
	ctx context.Context,
	userID string,
	orgName string,
) (RolePermissions, error) {
	// Step 1: Get current version number
	version, err := c.versionManager.GetVersion(ctx, orgName, userID)
	if err != nil {
		// Log error but continue with version 1 (fallback)
		version = 1
	}

	cacheKey := c.buildVersionedCacheKey(orgName, userID, version)

	// Step 2: Try cache first (with timeout)
	cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cached, err := c.redis.Get(cacheCtx, cacheKey).Result()
	if err == nil {
		// Cache hit - deserialize role permissions
		var rolePerms RolePermissions
		if jsonErr := json.Unmarshal([]byte(cached), &rolePerms); jsonErr != nil {
			// Continue to DB load on deserialization error
		} else {
			return rolePerms, nil
		}
	} else if err != redis.Nil {
		// Redis error (not a cache miss) - log and continue to DB
		logfacade.GetLogger(ctx).Warn(ctx, "Redis cache error, falling back to database", logfacade.Err(err))
	}

	// Step 3: Cache miss - load from database
	rolePerms, err := c.permLoader.LoadUserPermissionsAndRoles(ctx, userID, orgName)
	if err != nil {
		return nil, err
	}

	// Step 4: Store in cache with current version (best effort - don't fail if cache write fails)
	go c.cacheRolePermissions(cacheKey, rolePerms)

	return rolePerms, nil
}

// cachePermissions stores permissions in Redis cache (async, best effort)
func (c *PermissionCache) cachePermissions(cacheKey string, permissions []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := json.Marshal(permissions)
	if err != nil {
		return
	}

	c.redis.Set(ctx, cacheKey, data, c.cacheTTL)
}

// cacheRolePermissions stores role permissions in Redis cache (async, best effort)
func (c *PermissionCache) cacheRolePermissions(cacheKey string, rolePerms RolePermissions) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := json.Marshal(rolePerms)
	if err != nil {
		return
	}

	c.redis.Set(ctx, cacheKey, data, c.cacheTTL)
}

// buildCacheKey generates the Redis cache key for user permissions (legacy, kept for reference)
// Format: "auth:{orgName}:{userId}"
func (c *PermissionCache) buildCacheKey(orgName, userID string) string {
	return fmt.Sprintf("auth:%s:%s", orgName, userID)
}

// buildVersionedCacheKey generates the versioned Redis cache key for user permissions
// Phase 3 Format: "auth:{orgName}:{userId}:{version}"
func (c *PermissionCache) buildVersionedCacheKey(orgName, userID string, version int64) string {
	return fmt.Sprintf("auth:%s:%s:%d", orgName, userID, version)
}
