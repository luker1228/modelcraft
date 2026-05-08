package repository

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/cluster"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/config"
	"modelcraft/pkg/logfacade"
	"sync"
	"time"
)

// ClusterConnectionManager 集群连接管理器
type ClusterConnectionManager struct {
	connections  sync.Map // key: string, value: *sql.DB
	repo         cluster.DatabaseClusterRepository
	dbConfig     *config.DatabaseConfig // 连接池配置
	ticker       *time.Ticker           // 定时器
	stopChan     chan struct{}          // 停止信号
	lastSyncTime time.Time              // 上次同步时间
	logger       logfacade.Logger       // 日志记录器
	isRunning    bool                   // 是否正在运行
	runningMutex sync.RWMutex           // 运行状态锁
}

// DefaultClusterManager 默认集群连接管理器
var DefaultClusterManager *ClusterConnectionManager

// InitClusterManager 初始化集群连接管理器
func InitClusterManager(
	repo cluster.DatabaseClusterRepository,
	dbConfig *config.DatabaseConfig,
	logger logfacade.Logger,
) {
	DefaultClusterManager = newClusterConnectionManager(repo, dbConfig, logger)
}

// newClusterConnectionManager 创建新的集群连接管理器
func newClusterConnectionManager(
	repo cluster.DatabaseClusterRepository,
	dbConfig *config.DatabaseConfig,
	logger logfacade.Logger,
) *ClusterConnectionManager {
	return &ClusterConnectionManager{
		repo:         repo,
		dbConfig:     dbConfig,
		logger:       logger,
		stopChan:     make(chan struct{}),
		lastSyncTime: time.Now(),
	}
}

// GetConnection 获取数据库连接
func (cm *ClusterConnectionManager) GetConnection(
	ctx context.Context,
	orgName, projectSlug string,
) (*sql.DB, error) {
	clusterInfo, err := cm.repo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	clusterID := clusterInfo.ID

	// 先尝试从缓存获取
	if value, exists := cm.connections.Load(clusterID); exists {
		conn, ok := value.(*sql.DB)
		if !ok {
			return nil, bizerrors.New("invalid connection type in cache")
		}
		// 检查连接是否有效
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := conn.PingContext(pingCtx); err == nil {
			return conn, nil
		}
		// 连接无效，关闭并移除
		cm.closeConnection(clusterID)
	}

	connectionInfo := clusterInfo.GetConnectionInfo()
	// 创建新连接
	conn, err := cm.createConnection(connectionInfo)
	if err != nil {
		return nil, err
	}

	// 存储新连接到sync.Map
	cm.connections.Store(clusterID, conn)
	return conn, nil
}

func (cm *ClusterConnectionManager) TestConnection(ctx context.Context, connectionInfo cluster.ConnectionInfo) error {
	connection, err := cm.createConnection(connectionInfo)
	if err != nil {
		return err
	}
	defer func() { _ = connection.Close() }()

	timeout := time.Duration(connectionInfo.ConnectionTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = connection.PingContext(pingCtx)
	if err != nil {
		return err
	}
	return nil
}

// createConnection 创建新的数据库连接
func (cm *ClusterConnectionManager) createConnection(connectionInfo cluster.ConnectionInfo) (*sql.DB, error) {
	plainPassword, err := connectionInfo.Password.GetPlainPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get plain password: %w", err)
	}

	timeoutSeconds := connectionInfo.ConnectionTimeout
	if timeoutSeconds == 0 {
		timeoutSeconds = 5
	}

	// 构建DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/?charset=utf8mb4&collation=utf8mb4_unicode_ci&"+
			"parseTime=true&loc=Local&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		connectionInfo.Username,
		plainPassword,
		connectionInfo.Host,
		connectionInfo.Port,
		timeoutSeconds,
		timeoutSeconds,
		timeoutSeconds,
	)

	// 创建连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 设置连接池参数（使用配置文件中的设置）
	db.SetMaxOpenConns(cm.dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(cm.dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cm.dbConfig.ConnMaxLifetime) * time.Second)

	return db, nil
}

// RefreshConnection 刷新连接
func (cm *ClusterConnectionManager) RefreshConnection(
	ctx context.Context,
	orgName, projectSlug string,
) error {
	// 重新获取连接（这会触发重新获取 cluster 和新建连接）
	_, err := cm.GetConnection(ctx, orgName, projectSlug)
	return err
}

// closeConnection 关闭指定的连接
func (cm *ClusterConnectionManager) closeConnection(clusterName string) {
	if value, exists := cm.connections.LoadAndDelete(clusterName); exists {
		if conn, ok := value.(*sql.DB); ok {
			_ = conn.Close()
		}
	}
}

// CloseAll 关闭所有连接
func (cm *ClusterConnectionManager) CloseAll() error {
	// 首先停止定期同步任务
	cm.StopPeriodicSync()

	var errors []error
	cm.connections.Range(func(key, value interface{}) bool {
		clusterName, ok1 := key.(string)
		conn, ok2 := value.(*sql.DB)
		if !ok1 || !ok2 {
			return true
		}
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection for cluster %s: %w", clusterName, err))
		}
		cm.connections.Delete(key)
		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing connections: %v", errors)
	}

	cm.logger.Info(context.Background(), "所有集群连接已关闭")
	return nil
}

// GetConnectionWithDatabase 获取指定数据库的连接
func (cm *ClusterConnectionManager) GetConnectionWithDatabase(
	ctx context.Context,
	orgName, projectSlug, database string,
) (*sql.DB, error) {
	// 先获取基础连接
	conn, err := cm.GetConnection(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}

	// 切换到指定数据库
	// NOCA:yunding/go/sql-injection(设计如此)
	_, err = conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database))
	if err != nil {
		return nil, fmt.Errorf("failed to use database %s: %w", database, err)
	}

	return conn, nil
}

// HealthCheck 健康检查
func (cm *ClusterConnectionManager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	cm.connections.Range(func(key, value interface{}) bool {
		clusterName, ok1 := key.(string)
		conn, ok2 := value.(*sql.DB)
		if !ok1 || !ok2 {
			return true
		}
		if err := conn.PingContext(ctx); err != nil {
			results[clusterName] = err
		} else {
			results[clusterName] = nil
		}
		return true
	})

	return results
}

// StartPeriodicSync 启动定期同步任务
func (cm *ClusterConnectionManager) StartPeriodicSync(ctx context.Context) {
	cm.runningMutex.Lock()
	defer cm.runningMutex.Unlock()

	if cm.isRunning {
		cm.logger.Warn(context.Background(), "定期同步任务已在运行中")
		return
	}

	cm.ticker = time.NewTicker(1 * time.Minute)
	cm.isRunning = true

	cm.logger.Info(context.Background(), "启动集群连接池定期同步任务", logfacade.Duration("interval", 1*time.Minute))

	// 启动后台协程执行定期同步
	bizutils.GoWithCtx(ctx, func(ctx context.Context) {
		for {
			select {
			case <-cm.ticker.C:
				cm.syncConnections(ctx)
			case <-cm.stopChan:
				cm.logger.Info(context.Background(), "定期同步任务已停止")
				return
			}
		}
	})
}

// StopPeriodicSync 停止定期同步任务
func (cm *ClusterConnectionManager) StopPeriodicSync() {
	cm.runningMutex.Lock()
	defer cm.runningMutex.Unlock()

	if !cm.isRunning {
		return
	}

	if cm.ticker != nil {
		cm.ticker.Stop()
		cm.ticker = nil
	}

	close(cm.stopChan)
	cm.stopChan = make(chan struct{}) // 重新创建channel以便下次使用
	cm.isRunning = false

	cm.logger.Info(context.Background(), "定期同步任务已停止")
}

// syncConnections 同步连接池
func (cm *ClusterConnectionManager) syncConnections(ctx context.Context) {
	cm.logger.Debug(context.Background(),
		"开始同步集群连接池", logfacade.String("last_sync_time", cm.lastSyncTime.Format(time.RFC3339)))

	// 获取自上次同步以来更新的集群（传入空字符串表示跨所有组织和项目）
	updatedClusters, err := cm.repo.ListUpdatedAfter(
		ctx, cm.lastSyncTime, cluster.ClusterStatusActive,
	)
	if err != nil {
		cm.logger.Error(context.Background(), "获取更新的集群失败", logfacade.Err(err))
		return
	}

	if len(updatedClusters) == 0 {
		cm.logger.Debug(context.Background(), "没有发现更新的集群")
		cm.lastSyncTime = time.Now()
		return
	}

	cm.logger.Info(context.Background(), "发现更新的集群", logfacade.Int("count", len(updatedClusters)))

	// 更新连接池
	for _, clusterInfo := range updatedClusters {
		cm.updateConnection(ctx, clusterInfo)
	}

	// 更新同步时间
	cm.lastSyncTime = time.Now()
	cm.logger.Debug(
		context.Background(),
		"集群连接池同步完成",
		logfacade.String("sync_time", cm.lastSyncTime.Format(time.RFC3339)),
	)
}

// updateConnection 更新单个集群的连接
func (cm *ClusterConnectionManager) updateConnection(ctx context.Context, clusterInfo *cluster.DatabaseCluster) {
	clusterID := clusterInfo.ID

	cm.logger.Debug(context.Background(), "更新集群连接",
		logfacade.String("cluster_id", clusterID),
		logfacade.String("host", clusterInfo.Host),
		logfacade.Int("port", clusterInfo.Port))

	// 关闭现有连接
	if value, exists := cm.connections.LoadAndDelete(clusterID); exists {
		cm.closeExistingConnection(clusterID, value)
	}

	// 创建新连接
	connectionInfo := clusterInfo.GetConnectionInfo()
	newConn, err := cm.createConnection(connectionInfo)
	if err != nil {
		cm.logger.Error(context.Background(), "创建新连接失败",
			logfacade.String("cluster_id", clusterID),
			logfacade.Err(err))
		return
	}

	// 测试新连接
	pingTimeout := time.Duration(connectionInfo.ConnectionTimeout) * time.Second
	if pingTimeout <= 0 {
		pingTimeout = 5 * time.Second
	}
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := newConn.PingContext(pingCtx); err != nil {
		cm.logger.Error(context.Background(), "新连接测试失败",
			logfacade.String("cluster_id", clusterID),
			logfacade.Err(err))
		_ = newConn.Close()
		return
	}

	// 存储新连接
	cm.connections.Store(clusterID, newConn)
	cm.logger.Info(context.Background(), "集群连接已更新",
		logfacade.String("cluster_id", clusterID),
		logfacade.String("host", clusterInfo.Host),
		logfacade.Int("port", clusterInfo.Port))
}

func (cm *ClusterConnectionManager) closeExistingConnection(clusterName string, value interface{}) {
	conn, ok := value.(*sql.DB)
	if !ok {
		return
	}
	if err := conn.Close(); err != nil {
		cm.logger.Warn(context.Background(), "关闭旧连接失败",
			logfacade.String("cluster_name", clusterName),
			logfacade.Err(err))
		return
	}
	cm.logger.Debug(context.Background(), "已关闭旧连接", logfacade.String("cluster_name", clusterName))
}

// IsRunning 检查定期同步任务是否正在运行
func (cm *ClusterConnectionManager) IsRunning() bool {
	cm.runningMutex.RLock()
	defer cm.runningMutex.RUnlock()
	return cm.isRunning
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name string // 数据库名称
}

// ListDatabases 列出指定集群中的数据库（带分页）
func (cm *ClusterConnectionManager) ListDatabases(
	ctx context.Context,
	orgName, projectSlug, search string,
	offset, limit int,
) ([]*DatabaseInfo, int, error) {
	// 获取数据库连接
	conn, err := cm.GetConnection(ctx, orgName, projectSlug)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get connection: %w", err)
	}

	// 构建基础 WHERE 条件
	whereClause := "WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys', 'mc_meta')"
	args := []interface{}{}

	if search != "" {
		whereClause += " AND SCHEMA_NAME LIKE ?"
		args = append(args, "%"+search+"%")
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM information_schema.SCHEMATA " + whereClause
	var totalCount int
	if err := conn.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count databases: %w", err)
	}

	// 查询数据库名称（带分页）
	//nolint:gosec // whereClause is built from safe constants, not user input
	query := `
		SELECT SCHEMA_NAME
		FROM information_schema.SCHEMATA
		` + whereClause + `
		ORDER BY SCHEMA_NAME
		LIMIT ? OFFSET ?
	`
	queryArgs := make([]interface{}, 0, len(args)+2)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := conn.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	// 解析结果
	var databases []*DatabaseInfo
	for rows.Next() {
		var dbInfo DatabaseInfo
		if err := rows.Scan(&dbInfo.Name); err != nil {
			cm.logger.Warn(context.Background(), "failed to scan database info", logfacade.Err(err))
			continue
		}
		databases = append(databases, &dbInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate databases: %w", err)
	}

	cm.logger.Info(context.Background(), "successfully listed databases",
		logfacade.String("project", projectSlug),
		logfacade.Int("total", totalCount),
		logfacade.Int("returned", len(databases)))

	return databases, totalCount, nil
}
