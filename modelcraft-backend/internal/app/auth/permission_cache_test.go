package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPermissionLoader is a mock implementation of PermissionLoaderInterface
type MockPermissionLoader struct {
	mock.Mock
}

func (m *MockPermissionLoader) LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error) {
	args := m.Called(ctx, userID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPermissionLoader) LoadUserPermissionsAndRoles(
	ctx context.Context,
	userID string,
	orgName string,
) (RolePermissions, error) {
	args := m.Called(ctx, userID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(RolePermissions), args.Error(1)
}

// setupTestCache creates a test cache with miniredis (in-memory Redis mock)
func setupTestCache(t *testing.T) (*PermissionCache, *miniredis.Miniredis, *MockPermissionLoader) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create mock permission loader
	mockLoader := new(MockPermissionLoader)

	// Create version manager
	versionManager := NewPermissionVersionManager(redisClient)

	// Create cache with 5-minute TTL and version manager
	cache := NewPermissionCache(
		redisClient,
		mockLoader,
		versionManager,
		5*time.Minute,
	)

	return cache, mr, mockLoader
}

// TestPermissionCache_GetUserPermissions_CacheHit tests cache hit scenario
func TestPermissionCache_GetUserPermissions_CacheHit(t *testing.T) {
	t.Run("should return cached permissions on cache hit", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-123"
		orgName := "acme-corp"
		expectedPerms := []string{"model:read", "model:write", "cluster:manage"}

		// Pre-populate version and cache (Phase 3: versioned cache key)
		versionKey := "perm_version:acme-corp:user-123"
		cacheKey := "auth:acme-corp:user-123:1"
		mr.Set(versionKey, "1")
		data, _ := json.Marshal(expectedPerms)
		mr.Set(cacheKey, string(data))
		mr.SetTTL(cacheKey, 5*time.Minute)

		// MockLoader should NOT be called on cache hit
		mockLoader.AssertNotCalled(t, "LoadUserPermissions")

		// Act
		ctx := context.Background()
		start := time.Now()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)
		duration := time.Since(start)

		// Assert
		assert.NoError(t, err, "Should not return error on cache hit")
		assert.Equal(t, expectedPerms, result, "Should return cached permissions")
		assert.Less(t, duration, 10*time.Millisecond, "Cache hit should be < 10ms")

		// Verify mock was not called
		mockLoader.AssertExpectations(t)
	})

	t.Run("should return empty array from cache", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-456"
		orgName := "test-org"
		expectedPerms := []string{} // Empty permissions

		// Pre-populate version and cache with empty array (Phase 3: versioned cache key)
		versionKey := "perm_version:test-org:user-456"
		cacheKey := "auth:test-org:user-456:1"
		mr.Set(versionKey, "1")
		data, _ := json.Marshal(expectedPerms)
		mr.Set(cacheKey, string(data))

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPerms, result)
		assert.Empty(t, result, "Should return empty array")
		mockLoader.AssertNotCalled(t, "LoadUserPermissions")
	})
}

// TestPermissionCache_GetUserPermissions_CacheMiss tests cache miss scenario
func TestPermissionCache_GetUserPermissions_CacheMiss(t *testing.T) {
	t.Run("should load from DB and cache on cache miss", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-789"
		orgName := "new-org"
		expectedPerms := []string{"model:read", "project:create"}

		// Setup mock to return permissions from DB
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		// Act
		ctx := context.Background()
		start := time.Now()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)
		duration := time.Since(start)

		// Assert
		assert.NoError(t, err, "Should not return error")
		assert.Equal(t, expectedPerms, result, "Should return DB permissions")
		assert.Less(t, duration, 100*time.Millisecond, "Cache miss should be < 100ms")

		// Verify mock was called once
		mockLoader.AssertExpectations(t)
		mockLoader.AssertCalled(t, "LoadUserPermissions", mock.Anything, userID, orgName)

		// Wait for async cache write
		time.Sleep(150 * time.Millisecond)

		// Verify permissions were cached with version 1 (Phase 3: versioned cache key)
		cacheKey := "auth:new-org:user-789:1"
		exists := mr.Exists(cacheKey)
		assert.True(t, exists, "Permissions should be cached")

		cached, _ := mr.Get(cacheKey)
		var cachedPerms []string
		json.Unmarshal([]byte(cached), &cachedPerms)
		assert.Equal(t, expectedPerms, cachedPerms, "Cached permissions should match")
	})

	t.Run("should propagate DB errors", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-error"
		orgName := "error-org"
		expectedErr := assert.AnError

		// Setup mock to return error
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(nil, expectedErr)

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert
		assert.Error(t, err, "Should propagate DB error")
		assert.Nil(t, result, "Should return nil on error")
		assert.Equal(t, expectedErr, err, "Should return exact error from DB")

		// Verify mock was called
		mockLoader.AssertExpectations(t)
	})

	t.Run("should cache empty permissions array", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-no-perms"
		orgName := "no-perms-org"
		expectedPerms := []string{} // User has no permissions

		// Setup mock to return empty permissions
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPerms, result)
		assert.Empty(t, result)

		// Wait for async cache write
		time.Sleep(150 * time.Millisecond)

		// Verify empty array was cached (not nil) with version 1 (Phase 3: versioned cache key)
		cacheKey := "auth:no-perms-org:user-no-perms:1"
		exists := mr.Exists(cacheKey)
		assert.True(t, exists, "Empty permissions should be cached")

		cached, _ := mr.Get(cacheKey)
		var cachedPerms []string
		json.Unmarshal([]byte(cached), &cachedPerms)
		assert.NotNil(t, cachedPerms, "Cached value should not be nil")
		assert.Empty(t, cachedPerms, "Cached permissions should be empty array")
	})
}

// TestPermissionCache_CacheTTL tests cache expiration
func TestPermissionCache_CacheTTL(t *testing.T) {
	t.Run("should expire cache after TTL", func(t *testing.T) {
		// Arrange
		mr, err := miniredis.Run()
		require.NoError(t, err)
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		mockLoader := new(MockPermissionLoader)

		// Create version manager
		versionManager := NewPermissionVersionManager(redisClient)

		// Create cache with 1-second TTL for testing
		cache := NewPermissionCache(
			redisClient,
			mockLoader,
			versionManager,
			1*time.Second,
		)

		userID := "user-ttl"
		orgName := "ttl-org"
		expectedPerms := []string{"model:read"}

		// Setup mock for first call (cache miss)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil).Once()

		// Act 1: First call - cache miss
		ctx := context.Background()
		result1, err1 := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert 1: First call successful
		assert.NoError(t, err1)
		assert.Equal(t, expectedPerms, result1)

		// Wait for cache write
		time.Sleep(150 * time.Millisecond)

		// Act 2: Second call immediately - should hit cache
		result2, err2 := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert 2: Cache hit (mock called only once)
		assert.NoError(t, err2)
		assert.Equal(t, expectedPerms, result2)
		mockLoader.AssertNumberOfCalls(t, "LoadUserPermissions", 1)

		// Act 3: Wait for TTL expiration
		mr.FastForward(2 * time.Second) // Fast forward time in miniredis

		// Setup mock for second DB call (cache expired)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil).Once()

		// Act 4: Third call after TTL - should miss cache and reload from DB
		result3, err3 := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert 3: Cache expired, DB called again
		assert.NoError(t, err3)
		assert.Equal(t, expectedPerms, result3)
		mockLoader.AssertNumberOfCalls(t, "LoadUserPermissions", 2)
	})
}

// TestPermissionCache_RedisFailure tests graceful degradation on Redis failure
func TestPermissionCache_RedisFailure(t *testing.T) {
	t.Run("should fall back to DB when Redis is unavailable", func(t *testing.T) {
		// Arrange
		mr, err := miniredis.Run()
		require.NoError(t, err)

		redisClient := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})

		mockLoader := new(MockPermissionLoader)

		// Create version manager
		versionManager := NewPermissionVersionManager(redisClient)

		cache := NewPermissionCache(
			redisClient,
			mockLoader,
			versionManager,
			5*time.Minute,
		)

		// Simulate Redis failure by closing miniredis
		mr.Close()

		userID := "user-redis-fail"
		orgName := "redis-fail-org"
		expectedPerms := []string{"model:read"}

		// Setup mock to return permissions from DB
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert - should succeed despite Redis failure
		assert.NoError(t, err, "Should fall back to DB when Redis fails")
		assert.Equal(t, expectedPerms, result, "Should return DB permissions")

		// Verify DB was called
		mockLoader.AssertExpectations(t)
	})

	t.Run("should handle invalid JSON in cache gracefully", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-bad-json"
		orgName := "bad-json-org"
		expectedPerms := []string{"model:read"}

		// Pre-populate version and cache with invalid JSON (Phase 3: versioned cache key)
		versionKey := "perm_version:bad-json-org:user-bad-json"
		cacheKey := "auth:bad-json-org:user-bad-json:1"
		mr.Set(versionKey, "1")
		mr.Set(cacheKey, "invalid-json{{{")

		// Setup mock to return permissions from DB (fallback)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert - should fall back to DB on deserialization error
		assert.NoError(t, err, "Should fall back to DB on invalid JSON")
		assert.Equal(t, expectedPerms, result, "Should return DB permissions")

		// Verify DB was called
		mockLoader.AssertExpectations(t)
	})
}

// TestPermissionCache_BuildCacheKey tests versioned cache key generation (Phase 3)
func TestPermissionCache_BuildCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		userID   string
		version  int64
		expected string
	}{
		{
			name:     "standard versioned key",
			orgName:  "acme-corp",
			userID:   "user-123",
			version:  1,
			expected: "auth:acme-corp:user-123:1",
		},
		{
			name:     "key with special characters",
			orgName:  "test_org-2024",
			userID:   "user-uuid-456-789",
			version:  5,
			expected: "auth:test_org-2024:user-uuid-456-789:5",
		},
		{
			name:     "key with uppercase and higher version",
			orgName:  "ACME",
			userID:   "USER-XYZ",
			version:  42,
			expected: "auth:ACME:USER-XYZ:42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cache, mr, _ := setupTestCache(t)
			defer mr.Close()

			// Act - test buildVersionedCacheKey (Phase 3)
			result := cache.buildVersionedCacheKey(tt.orgName, tt.userID, tt.version)

			// Assert
			assert.Equal(t, tt.expected, result, "Versioned cache key should match expected format")
		})
	}

	// Test legacy buildCacheKey still works (for backward compatibility)
	t.Run("legacy buildCacheKey without version", func(t *testing.T) {
		cache, mr, _ := setupTestCache(t)
		defer mr.Close()

		result := cache.buildCacheKey("test-org", "user-123")
		assert.Equal(t, "auth:test-org:user-123", result, "Legacy cache key should not include version")
	})
}

// TestPermissionCache_PerformanceBenchmark benchmarks cache performance
func TestPermissionCache_PerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("cache hit should be < 1ms", func(t *testing.T) {
		// Arrange
		cache, mr, _ := setupTestCache(t)
		defer mr.Close()

		userID := "user-perf"
		orgName := "perf-org"
		expectedPerms := []string{"model:read", "model:write", "cluster:manage"}

		// Pre-populate version and cache (Phase 3: versioned cache key)
		versionKey := "perm_version:perf-org:user-perf"
		cacheKey := "auth:perf-org:user-perf:1"
		mr.Set(versionKey, "1")
		data, _ := json.Marshal(expectedPerms)
		mr.Set(cacheKey, string(data))

		ctx := context.Background()

		// Warm up
		cache.GetUserPermissions(ctx, userID, orgName)

		// Measure 100 cache hits
		var totalDuration time.Duration
		iterations := 100

		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := cache.GetUserPermissions(ctx, userID, orgName)
			duration := time.Since(start)

			assert.NoError(t, err)
			totalDuration += duration
		}

		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("Average cache hit latency: %v", avgDuration)

		// Assert average < 1ms
		assert.Less(t, avgDuration, 1*time.Millisecond,
			"Average cache hit should be < 1ms, got %v", avgDuration)
	})

	t.Run("cache miss should be < 50ms", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-perf-miss"
		orgName := "perf-miss-org"
		expectedPerms := []string{"model:read"}

		// Setup mock with fast response (simulating DB query)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		ctx := context.Background()

		// Measure cache miss
		start := time.Now()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)
		duration := time.Since(start)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPerms, result)
		assert.Less(t, duration, 50*time.Millisecond,
			"Cache miss should be < 50ms, got %v", duration)

		t.Logf("Cache miss latency: %v", duration)
	})
}

// TestPermissionCache_ConcurrentAccess tests thread safety
func TestPermissionCache_ConcurrentAccess(t *testing.T) {
	t.Run("should handle concurrent requests safely", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-concurrent"
		orgName := "concurrent-org"
		expectedPerms := []string{"model:read"}

		// Setup mock to return permissions
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil)

		ctx := context.Background()

		// Act - 50 concurrent requests
		results := make(chan []string, 50)
		errors := make(chan error, 50)

		for i := 0; i < 50; i++ {
			go func() {
				perms, err := cache.GetUserPermissions(ctx, userID, orgName)
				if err != nil {
					errors <- err
				} else {
					results <- perms
				}
			}()
		}

		// Collect results
		var successCount int
		for i := 0; i < 50; i++ {
			select {
			case perms := <-results:
				assert.Equal(t, expectedPerms, perms)
				successCount++
			case err := <-errors:
				assert.NoError(t, err, "Should not have errors")
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// Assert all requests succeeded
		assert.Equal(t, 50, successCount, "All concurrent requests should succeed")
	})
}

// TestPermissionCache_VersionedCacheKeys tests Phase 3 versioned cache behavior
func TestPermissionCache_VersionedCacheKeys(t *testing.T) {
	t.Run("should include version in cache key", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-versioned"
		orgName := "versioned-org"
		expectedPerms := []string{"model:read", "model:write"}

		// Setup mock for first load
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil).Once()

		// Act - first call (cache miss, version 1)
		ctx := context.Background()
		result1, err1 := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert - first call loaded from DB
		assert.NoError(t, err1)
		assert.Equal(t, expectedPerms, result1)

		// Wait for async cache write
		time.Sleep(150 * time.Millisecond)

		// Verify cache key includes version 1
		versionedKey := "auth:versioned-org:user-versioned:1"
		exists := mr.Exists(versionedKey)
		assert.True(t, exists, "Cache key should include version number")

		cached, _ := mr.Get(versionedKey)
		var cachedPerms []string
		json.Unmarshal([]byte(cached), &cachedPerms)
		assert.Equal(t, expectedPerms, cachedPerms)
	})

	t.Run("should use version 1 by default on first access", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-new"
		orgName := "new-org"
		expectedPerms := []string{"model:read"}

		// Setup mock
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(expectedPerms, nil).Once()

		// Act
		ctx := context.Background()
		result, err := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPerms, result)

		// Wait for cache write
		time.Sleep(150 * time.Millisecond)

		// Verify version 1 key exists
		versionKey := "perm_version:new-org:user-new"
		cacheKey := "auth:new-org:user-new:1"

		time.Sleep(150 * time.Millisecond) // Wait for version initialization

		versionExists := mr.Exists(versionKey)
		cacheExists := mr.Exists(cacheKey)

		assert.True(t, versionExists, "Version key should be initialized")
		assert.True(t, cacheExists, "Cache key should use version 1")
	})
}

// TestPermissionCache_VersionMismatchInvalidation tests cache invalidation via version increment
func TestPermissionCache_VersionMismatchInvalidation(t *testing.T) {
	t.Run("should invalidate cache when version increments", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-invalidate"
		orgName := "invalidate-org"
		oldPerms := []string{"model:read"}
		newPerms := []string{"model:read", "model:write", "cluster:manage"}

		// Setup mock for first load (old permissions)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(oldPerms, nil).Once()

		ctx := context.Background()

		// Act 1: First call - load and cache with version 1
		result1, err1 := cache.GetUserPermissions(ctx, userID, orgName)
		assert.NoError(t, err1)
		assert.Equal(t, oldPerms, result1)

		// Wait for cache write
		time.Sleep(150 * time.Millisecond)

		// Verify old cache exists with version 1
		oldCacheKey := "auth:invalidate-org:user-invalidate:1"
		assert.True(t, mr.Exists(oldCacheKey), "Old cache key should exist")

		// Act 2: Simulate permission change - increment version
		versionKey := "perm_version:invalidate-org:user-invalidate"
		mr.Incr(versionKey, 1) // Version becomes 2

		// Setup mock for second load (new permissions)
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(newPerms, nil).Once()

		// Act 3: Second call - should miss cache (version mismatch) and reload
		result2, err2 := cache.GetUserPermissions(ctx, userID, orgName)

		// Assert - new permissions loaded
		assert.NoError(t, err2)
		assert.Equal(t, newPerms, result2, "Should load new permissions after version increment")

		// Wait for new cache write
		time.Sleep(150 * time.Millisecond)

		// Verify new cache exists with version 2
		newCacheKey := "auth:invalidate-org:user-invalidate:2"
		assert.True(t, mr.Exists(newCacheKey), "New cache key should exist with version 2")

		// Verify DB was called twice (once for each version)
		mockLoader.AssertNumberOfCalls(t, "LoadUserPermissions", 2)
	})

	t.Run("should load fresh permissions after multiple version increments", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-multi-version"
		orgName := "multi-version-org"
		perms1 := []string{"model:read"}
		perms2 := []string{"model:read", "model:write"}
		perms3 := []string{"model:read", "model:write", "cluster:manage"}

		ctx := context.Background()

		// Setup mock for version 1
		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(perms1, nil).Once()

		// Act 1: Load with version 1
		result1, err1 := cache.GetUserPermissions(ctx, userID, orgName)
		assert.NoError(t, err1)
		assert.Equal(t, perms1, result1)
		time.Sleep(150 * time.Millisecond)

		// Act 2: Increment to version 2
		versionKey := "perm_version:multi-version-org:user-multi-version"
		mr.Incr(versionKey, 1)

		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(perms2, nil).Once()

		result2, err2 := cache.GetUserPermissions(ctx, userID, orgName)
		assert.NoError(t, err2)
		assert.Equal(t, perms2, result2)
		time.Sleep(150 * time.Millisecond)

		// Act 3: Increment to version 3
		mr.Incr(versionKey, 1)

		mockLoader.On("LoadUserPermissions", mock.Anything, userID, orgName).
			Return(perms3, nil).Once()

		result3, err3 := cache.GetUserPermissions(ctx, userID, orgName)
		assert.NoError(t, err3)
		assert.Equal(t, perms3, result3)

		// Assert - DB called 3 times (once per version)
		mockLoader.AssertNumberOfCalls(t, "LoadUserPermissions", 3)
	})
}

// TestPermissionCache_VersionedPerformance tests performance with versioning
func TestPermissionCache_VersionedPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("versioned cache hit should still be < 1ms", func(t *testing.T) {
		// Arrange
		cache, mr, mockLoader := setupTestCache(t)
		defer mr.Close()

		userID := "user-perf-versioned"
		orgName := "perf-versioned-org"
		expectedPerms := []string{"model:read", "model:write", "cluster:manage"}

		// Pre-populate version and cache
		versionKey := "perm_version:perf-versioned-org:user-perf-versioned"
		cacheKey := "auth:perf-versioned-org:user-perf-versioned:1"
		mr.Set(versionKey, "1")
		data, _ := json.Marshal(expectedPerms)
		mr.Set(cacheKey, string(data))

		ctx := context.Background()

		// Warm up
		cache.GetUserPermissions(ctx, userID, orgName)

		// Measure 100 cache hits
		var totalDuration time.Duration
		iterations := 100

		for i := 0; i < iterations; i++ {
			start := time.Now()
			result, err := cache.GetUserPermissions(ctx, userID, orgName)
			duration := time.Since(start)

			assert.NoError(t, err)
			assert.Equal(t, expectedPerms, result)
			totalDuration += duration
		}

		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("Average versioned cache hit latency: %v", avgDuration)

		// Assert average < 1ms (same performance target as non-versioned)
		assert.Less(t, avgDuration, 1*time.Millisecond,
			"Versioned cache hit should still be < 1ms, got %v", avgDuration)

		// Verify DB was never called (all cache hits)
		mockLoader.AssertNotCalled(t, "LoadUserPermissions")
	})
}
