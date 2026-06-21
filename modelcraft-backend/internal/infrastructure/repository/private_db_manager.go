package repository

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/config"
	"modelcraft/pkg/logfacade"
	"sync"
	"time"
)

// PrivateDBManager manages private database connections for end-user auth.
//
// Deprecated: Cache key uses orgName.projectSlug but EndUser is now Org-scoped.
// The projectSlug dimension will be removed in a future cleanup — connections
// should be keyed by orgName only.
//
// Target DB: mc_meta
//
// Behavior:
// 1. Return cached healthy connection
// 2. Rebuild when connection invalid
// 3. Keep one dedicated *sql.DB per project key
//
// Note:
// It reuses ClusterConnectionManager's repository in the same package,
// but does not reuse ClusterConnectionManager's shared *sql.DB instance to avoid
// cross-project database switching side effects.
type PrivateDBManager struct {
	connections sync.Map // key: orgName.projectSlug -> *sql.DB

	clusterMgr *ClusterConnectionManager
	dbConfig   *config.DatabaseConfig
	logger     logfacade.Logger

	mu sync.Mutex
}

const endUserMetadataDBName = "mc_meta"

// NewPrivateDBManager creates a new PrivateDBManager.
func NewPrivateDBManager(
	clusterMgr *ClusterConnectionManager,
	dbConfig *config.DatabaseConfig,
	logger logfacade.Logger,
) *PrivateDBManager {
	return &PrivateDBManager{
		clusterMgr: clusterMgr,
		dbConfig:   dbConfig,
		logger:     logger,
	}
}

func privateCacheKey(orgName, projectSlug string) string {
	return orgName + "." + projectSlug
}

// GetOrInit gets a private DB connection.
func (m *PrivateDBManager) GetOrInit(ctx context.Context, orgName, projectSlug string) (*sql.DB, error) {
	key := privateCacheKey(orgName, projectSlug)

	if conn := m.getHealthyFromCache(ctx, key); conn != nil {
		return conn, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if conn := m.getHealthyFromCache(ctx, key); conn != nil {
		return conn, nil
	}

	return m.createAndCacheConnection(ctx, orgName, projectSlug, key)
}

func (m *PrivateDBManager) getHealthyFromCache(ctx context.Context, key string) *sql.DB {
	value, exists := m.connections.Load(key)
	if !exists {
		return nil
	}

	conn, ok := value.(*sql.DB)
	if !ok {
		m.logger.With(logfacade.String("key", key)).Warnf(ctx, "private DB cache type invalid, evicting")
		m.connections.Delete(key)
		return nil
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	err := conn.PingContext(pingCtx)
	cancel()
	if err == nil {
		return conn
	}

	m.logger.With(
		logfacade.String("key", key),
		logfacade.Err(err)).Warnf(ctx, "private DB ping failed, rebuilding")
	m.evictByKey(key)
	return nil
}

func (m *PrivateDBManager) createAndCacheConnection(
	ctx context.Context,
	orgName, projectSlug, key string,
) (*sql.DB, error) {
	connectionInfo, plainPassword, err := m.getClusterConnectionInfo(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}

	dbName := endUserMetadataDBName
	privateDB, err := m.openPrivateDBConnection(connectionInfo, plainPassword, dbName)
	if err != nil {
		return nil, fmt.Errorf("open private DB connection: %w", err)
	}

	m.connections.Store(key, privateDB)
	m.logger.With(
		logfacade.String("key", key),
		logfacade.String("database", dbName)).Infof(ctx, "private DB connected")

	return privateDB, nil
}

func (m *PrivateDBManager) getClusterConnectionInfo(
	ctx context.Context,
	orgName, projectSlug string,
) (cluster.ConnectionInfo, string, error) {
	if m.clusterMgr == nil || m.clusterMgr.repo == nil {
		return cluster.ConnectionInfo{}, "", fmt.Errorf("cluster manager not initialized")
	}

	clusterInfo, err := m.clusterMgr.repo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return cluster.ConnectionInfo{}, "", shared.NewNotFoundError(
				fmt.Sprintf("cluster not configured for project: %s/%s", orgName, projectSlug),
			)
		}
		return cluster.ConnectionInfo{}, "", fmt.Errorf("get cluster by project key: %w", err)
	}

	connInfo := clusterInfo.GetConnectionInfo()
	plainPassword, err := connInfo.Password.GetPlainPassword()
	if err != nil {
		return cluster.ConnectionInfo{}, "", fmt.Errorf("decrypt cluster password: %w", err)
	}
	return connInfo, plainPassword, nil
}

func (m *PrivateDBManager) openPrivateDBConnection(
	connectionInfo cluster.ConnectionInfo,
	plainPassword, database string,
) (*sql.DB, error) {
	timeoutSeconds := connectionInfo.ConnectionTimeout
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}

	dsn := buildPrivateDSN(
		connectionInfo.Username,
		plainPassword,
		connectionInfo.Host,
		connectionInfo.Port,
		database,
		timeoutSeconds,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if m.dbConfig != nil {
		db.SetMaxOpenConns(m.dbConfig.MaxOpenConns)
		db.SetMaxIdleConns(m.dbConfig.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(m.dbConfig.ConnMaxLifetime) * time.Second)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()
	if err = db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func buildPrivateDSN(username, password, host string, port int, database string, timeoutSeconds int) string {
	dsnSuffix := "?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local"
	timeoutSuffix := fmt.Sprintf(
		"&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		timeoutSeconds,
		timeoutSeconds,
		timeoutSeconds,
	)
	if database == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s", username, password, host, port, dsnSuffix, timeoutSuffix)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s%s", username, password, host, port, database, dsnSuffix, timeoutSuffix)
}

func (m *PrivateDBManager) evictByKey(key string) {
	if value, exists := m.connections.LoadAndDelete(key); exists {
		if db, ok := value.(*sql.DB); ok {
			_ = db.Close()
		}
	}
}

// EvictCache evicts one project's connection cache.
func (m *PrivateDBManager) EvictCache(orgName, projectSlug string) {
	m.evictByKey(privateCacheKey(orgName, projectSlug))
}

// Provision is kept for compatibility and now only warms mc_meta connection cache.
// Implements project.PrivateDBProvisioner.
func (m *PrivateDBManager) Provision(ctx context.Context, orgName, projectSlug string) error {
	key := privateCacheKey(orgName, projectSlug)

	m.mu.Lock()
	defer m.mu.Unlock()

	connectionInfo, plainPassword, err := m.getClusterConnectionInfo(ctx, orgName, projectSlug)
	if err != nil {
		return err
	}

	dbName := endUserMetadataDBName
	m.evictByKey(key)
	privateDB, err := m.openPrivateDBConnection(connectionInfo, plainPassword, dbName)
	if err != nil {
		return fmt.Errorf("open private DB connection: %w", err)
	}

	m.connections.Store(key, privateDB)
	m.logger.With(
		logfacade.String("key", key),
		logfacade.String("database", dbName)).Infof(ctx, "private DB provisioned")
	return nil
}

// CloseAll closes all cached private DB connections.
func (m *PrivateDBManager) CloseAll() {
	m.connections.Range(func(key, _ interface{}) bool {
		if k, ok := key.(string); ok {
			m.evictByKey(k)
		}
		return true
	})
}

// HealthCheck checks all cached private DB connections.
func (m *PrivateDBManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	m.connections.Range(func(key, value interface{}) bool {
		k, ok1 := key.(string)
		db, ok2 := value.(*sql.DB)
		if !ok1 || !ok2 {
			return true
		}
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := db.PingContext(pingCtx)
		cancel()
		results[k] = err
		return true
	})
	return results
}
