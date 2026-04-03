package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupVersionManager creates a test version manager with miniredis
func setupVersionManager(t *testing.T) (*PermissionVersionManager, *miniredis.Miniredis) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create version manager
	manager := NewPermissionVersionManager(redisClient)

	return manager, mr
}

// TestPermissionVersionManager_GetVersion tests getting version numbers
func TestPermissionVersionManager_GetVersion(t *testing.T) {
	t.Run("should return 1 on first access", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "acme-corp"
		userID := "user-123"

		// Act
		ctx := context.Background()
		version, err := manager.GetVersion(ctx, orgName, userID)

		// Assert
		assert.NoError(t, err, "Should not return error on first access")
		assert.Equal(t, int64(1), version, "Version should be 1 on first access")

		// Wait for async initialization
		time.Sleep(150 * time.Millisecond)

		// Verify version key was initialized in Redis
		versionKey := "perm_version:acme-corp:user-123"
		exists := mr.Exists(versionKey)
		assert.True(t, exists, "Version key should be initialized")

		storedVersion, _ := mr.Get(versionKey)
		assert.Equal(t, "1", storedVersion, "Stored version should be 1")
	})

	t.Run("should return existing version", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "test-org"
		userID := "user-456"
		versionKey := "perm_version:test-org:user-456"

		// Pre-set version to 5
		mr.Set(versionKey, "5")

		// Act
		ctx := context.Background()
		version, err := manager.GetVersion(ctx, orgName, userID)

		// Assert
		assert.NoError(t, err, "Should not return error")
		assert.Equal(t, int64(5), version, "Should return existing version")
	})

	t.Run("should handle Redis timeout gracefully", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		mr.Close() // Close Redis to simulate timeout

		orgName := "timeout-org"
		userID := "user-timeout"

		// Act
		ctx := context.Background()
		version, err := manager.GetVersion(ctx, orgName, userID)

		// Assert
		assert.Error(t, err, "Should return error when Redis is unavailable")
		assert.Equal(t, int64(0), version, "Should return 0 on error")
		assert.Contains(t, err.Error(), "failed to get permission version")
	})

	t.Run("should complete within 100ms timeout", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "perf-org"
		userID := "user-perf"

		// Act
		ctx := context.Background()
		start := time.Now()
		version, err := manager.GetVersion(ctx, orgName, userID)
		duration := time.Since(start)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(1), version)
		assert.Less(t, duration, 100*time.Millisecond,
			"GetVersion should complete within 100ms timeout")
	})
}

// TestPermissionVersionManager_IncrementVersion tests incrementing versions
func TestPermissionVersionManager_IncrementVersion(t *testing.T) {
	t.Run("should increment version from 1 to 2", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "acme-corp"
		userID := "user-789"
		versionKey := "perm_version:acme-corp:user-789"

		// Pre-set version to 1
		mr.Set(versionKey, "1")

		// Act
		ctx := context.Background()
		newVersion, err := manager.IncrementVersion(ctx, orgName, userID)

		// Assert
		assert.NoError(t, err, "Should not return error")
		assert.Equal(t, int64(2), newVersion, "Should increment to 2")

		// Verify in Redis
		storedVersion, _ := mr.Get(versionKey)
		assert.Equal(t, "2", storedVersion, "Stored version should be 2")
	})

	t.Run("should initialize to 1 on first increment", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "new-org"
		userID := "user-new"
		versionKey := "perm_version:new-org:user-new"

		// Verify key doesn't exist
		exists := mr.Exists(versionKey)
		assert.False(t, exists, "Version key should not exist initially")

		// Act
		ctx := context.Background()
		newVersion, err := manager.IncrementVersion(ctx, orgName, userID)

		// Assert
		assert.NoError(t, err, "Should not return error")
		assert.Equal(t, int64(1), newVersion, "Should initialize to 1 on first increment")

		// Verify in Redis
		storedVersion, _ := mr.Get(versionKey)
		assert.Equal(t, "1", storedVersion, "Stored version should be 1")
	})

	t.Run("should increment multiple times correctly", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "multi-org"
		userID := "user-multi"

		// Act - increment 5 times
		ctx := context.Background()
		var lastVersion int64
		for i := 1; i <= 5; i++ {
			version, err := manager.IncrementVersion(ctx, orgName, userID)
			assert.NoError(t, err)
			assert.Equal(t, int64(i), version, "Should increment sequentially")
			lastVersion = version
		}

		// Assert final version
		assert.Equal(t, int64(5), lastVersion, "Final version should be 5")
	})

	t.Run("should handle Redis error gracefully", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		mr.Close() // Close Redis to simulate error

		orgName := "error-org"
		userID := "user-error"

		// Act
		ctx := context.Background()
		version, err := manager.IncrementVersion(ctx, orgName, userID)

		// Assert
		assert.Error(t, err, "Should return error when Redis is unavailable")
		assert.Equal(t, int64(0), version, "Should return 0 on error")
		assert.Contains(t, err.Error(), "failed to increment permission version")
	})

	t.Run("should complete within 100ms timeout", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "perf-incr-org"
		userID := "user-perf-incr"

		// Act
		ctx := context.Background()
		start := time.Now()
		version, err := manager.IncrementVersion(ctx, orgName, userID)
		duration := time.Since(start)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(1), version)
		assert.Less(t, duration, 100*time.Millisecond,
			"IncrementVersion should complete within 100ms timeout")
	})
}

// TestPermissionVersionManager_VersionPersistence tests version persistence
func TestPermissionVersionManager_VersionPersistence(t *testing.T) {
	t.Run("should persist version without TTL", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "persist-org"
		userID := "user-persist"
		versionKey := "perm_version:persist-org:user-persist"

		// Act - increment version
		ctx := context.Background()
		_, err := manager.IncrementVersion(ctx, orgName, userID)
		require.NoError(t, err)

		// Assert - verify NO TTL is set (TTL = 0 means no expiration)
		ttl := mr.TTL(versionKey)
		assert.Equal(t, time.Duration(0), ttl, "Version key should have no TTL (persistent)")
	})

	t.Run("should maintain version across multiple operations", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "stable-org"
		userID := "user-stable"

		ctx := context.Background()

		// Act 1: Increment version
		v1, err1 := manager.IncrementVersion(ctx, orgName, userID)
		assert.NoError(t, err1)
		assert.Equal(t, int64(1), v1)

		// Act 2: Get version
		v2, err2 := manager.GetVersion(ctx, orgName, userID)
		assert.NoError(t, err2)
		assert.Equal(t, int64(1), v2, "Version should remain 1")

		// Act 3: Increment again
		v3, err3 := manager.IncrementVersion(ctx, orgName, userID)
		assert.NoError(t, err3)
		assert.Equal(t, int64(2), v3)

		// Act 4: Get version again
		v4, err4 := manager.GetVersion(ctx, orgName, userID)
		assert.NoError(t, err4)
		assert.Equal(t, int64(2), v4, "Version should now be 2")

		// Assert - version persists correctly
		assert.Equal(t, v3, v4, "Version should persist after increment")
	})
}

// TestPermissionVersionManager_ConcurrentIncrements tests thread safety
func TestPermissionVersionManager_ConcurrentIncrements(t *testing.T) {
	t.Run("should handle concurrent increments atomically", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "concurrent-org"
		userID := "user-concurrent"

		// Act - 100 concurrent increments
		ctx := context.Background()
		var wg sync.WaitGroup
		results := make(chan int64, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				version, err := manager.IncrementVersion(ctx, orgName, userID)
				if err != nil {
					t.Errorf("Increment failed: %v", err)
					return
				}
				results <- version
			}()
		}

		wg.Wait()
		close(results)

		// Collect all version numbers
		versions := make([]int64, 0, 100)
		for v := range results {
			versions = append(versions, v)
		}

		// Assert - should have exactly 100 unique sequential versions
		assert.Len(t, versions, 100, "Should have 100 versions")

		// Verify final version is 100
		versionKey := "perm_version:concurrent-org:user-concurrent"
		storedVersion, _ := mr.Get(versionKey)
		assert.Equal(t, "100", storedVersion, "Final version should be 100")

		// Verify all versions are unique (1 to 100)
		versionSet := make(map[int64]bool)
		for _, v := range versions {
			assert.False(t, versionSet[v], "Version %d should be unique", v)
			versionSet[v] = true
			assert.GreaterOrEqual(t, v, int64(1), "Version should be >= 1")
			assert.LessOrEqual(t, v, int64(100), "Version should be <= 100")
		}
	})

	t.Run("should isolate versions between different users", func(t *testing.T) {
		// Arrange
		manager, mr := setupVersionManager(t)
		defer mr.Close()

		orgName := "multi-user-org"
		user1 := "user-1"
		user2 := "user-2"

		ctx := context.Background()

		// Act - increment both users concurrently
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				manager.IncrementVersion(ctx, orgName, user1)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				manager.IncrementVersion(ctx, orgName, user2)
			}
		}()

		wg.Wait()

		// Assert - each user has independent version counter
		v1, err1 := manager.GetVersion(ctx, orgName, user1)
		assert.NoError(t, err1)
		assert.Equal(t, int64(10), v1, "User 1 should have version 10")

		v2, err2 := manager.GetVersion(ctx, orgName, user2)
		assert.NoError(t, err2)
		assert.Equal(t, int64(5), v2, "User 2 should have version 5")

		// Verify in Redis
		key1 := "perm_version:multi-user-org:user-1"
		key2 := "perm_version:multi-user-org:user-2"
		stored1, err := mr.Get(key1)
		require.NoError(t, err)
		stored2, err := mr.Get(key2)
		require.NoError(t, err)
		assert.Equal(t, "10", stored1)
		assert.Equal(t, "5", stored2)
	})
}

// TestPermissionVersionManager_BuildVersionKey tests key generation
func TestPermissionVersionManager_BuildVersionKey(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		userID   string
		expected string
	}{
		{
			name:     "standard key",
			orgName:  "acme-corp",
			userID:   "user-123",
			expected: "perm_version:acme-corp:user-123",
		},
		{
			name:     "key with special characters",
			orgName:  "test_org-2024",
			userID:   "user-uuid-456-789",
			expected: "perm_version:test_org-2024:user-uuid-456-789",
		},
		{
			name:     "key with uppercase",
			orgName:  "ACME",
			userID:   "USER-XYZ",
			expected: "perm_version:ACME:USER-XYZ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			manager, mr := setupVersionManager(t)
			defer mr.Close()

			// Act
			result := manager.buildVersionKey(tt.orgName, tt.userID)

			// Assert
			assert.Equal(t, tt.expected, result, "Version key should match expected format")
		})
	}
}
