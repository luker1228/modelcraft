package cluster

import (
	"modelcraft/pkg/bizerrors"
	"time"
)

// ClusterLocator 集群定位器
type ClusterLocator struct {
	OrgName     string `json:"orgName"`
	ProjectSlug string `json:"projectSlug"`
}

// Validate 验证集群定位器的必填字段
func (l *ClusterLocator) Validate() error {
	if l.OrgName == "" {
		return bizerrors.Errorf("OrgName cant be blank")
	}
	if l.ProjectSlug == "" {
		return bizerrors.Errorf("ProjectSlug cant be blank")
	}
	return nil
}

// NewClusterLocator 创建集群定位器并验证必填字段
func NewClusterLocator(orgName, projectSlug string) (ClusterLocator, error) {
	locator := ClusterLocator{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	}
	if err := locator.Validate(); err != nil {
		return locator, err
	}
	return locator, nil
}

// GetFullPath 获取集群完整路径
// 返回格式: org_name.project_slug
func (l *ClusterLocator) GetFullPath() string {
	return l.OrgName + "." + l.ProjectSlug
}

// DatabaseCluster 数据库集群实体
type DatabaseCluster struct {
	ID                string        `json:"id"`
	OrgName           string        `json:"orgName"`
	ProjectSlug       string        `json:"projectSlug"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	Host              string        `json:"host"`
	Port              int           `json:"port"`
	Username          string        `json:"username"`
	Password          Password      `json:"password"`          // 使用Password类型
	ConnectionTimeout int           `json:"connectionTimeout"` // Connection timeout in seconds (5-15, default 5)
	Status            ClusterStatus `json:"status"`
	Version           int64         `json:"version"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

// ConnectionInfo 数据库连接信息
type ConnectionInfo struct {
	Host              string   `json:"host"`
	Port              int      `json:"port"`
	Username          string   `json:"username"`
	Password          Password `json:"password"`          // 使用Password类型
	ConnectionTimeout int      `json:"connectionTimeout"` // Connection timeout in seconds
}

// ClusterStatus 集群状态
type ClusterStatus string

const (
	ClusterStatusActive   ClusterStatus = "active"
	ClusterStatusDisabled ClusterStatus = "disabled"
)

// Validate 验证数据库集群实体
func (dc *DatabaseCluster) Validate() error {
	if dc.OrgName == "" {
		return bizerrors.Errorf("org name cannot be empty")
	}
	if dc.ProjectSlug == "" {
		return bizerrors.Errorf("project slug cannot be empty")
	}
	if dc.Title == "" {
		return bizerrors.Errorf("cluster title cannot be empty")
	}
	if dc.Host == "" {
		return bizerrors.Errorf("host cannot be empty")
	}
	if dc.Port <= 0 || dc.Port > 65535 {
		return bizerrors.Errorf("port must be between 1 and 65535")
	}
	if dc.Username == "" {
		return bizerrors.Errorf("username cannot be empty")
	}
	if dc.ConnectionTimeout < 5 || dc.ConnectionTimeout > 15 {
		return bizerrors.Errorf("connection timeout must be between 5 and 15 seconds")
	}

	return nil
}

// GetClusterLocator 获取集群定位器
func (dc *DatabaseCluster) GetClusterLocator() *ClusterLocator {
	return &ClusterLocator{
		OrgName:     dc.OrgName,
		ProjectSlug: dc.ProjectSlug,
	}
}

// IsActive 检查集群是否处于活跃状态
func (dc *DatabaseCluster) IsActive() bool {
	return dc.Status == ClusterStatusActive
}

// SetDefaults 设置默认值
func (dc *DatabaseCluster) SetDefaults() {
	if dc.Port == 0 {
		dc.Port = 3306
	}
	if dc.ConnectionTimeout == 0 {
		dc.ConnectionTimeout = 5
	}

	if dc.Status == "" {
		dc.Status = ClusterStatusActive
	}
	if dc.Version == 0 {
		dc.Version = 1
	}
}

// UpdateConnectionTimeout 更新连接超时（单位：秒）
func (dc *DatabaseCluster) UpdateConnectionTimeout(timeout int) error {
	old := dc.ConnectionTimeout
	dc.ConnectionTimeout = timeout
	dc.UpdatedAt = time.Now()

	if err := dc.Validate(); err != nil {
		dc.ConnectionTimeout = old
		return err
	}
	return nil
}

// GetConnectionInfo 获取数据库连接信息
func (dc *DatabaseCluster) GetConnectionInfo() ConnectionInfo {
	return ConnectionInfo{
		Username:          dc.Username,
		Password:          dc.Password,
		Host:              dc.Host,
		Port:              dc.Port,
		ConnectionTimeout: dc.ConnectionTimeout,
	}
}

// NewDatabaseCluster 创建新的数据库集群实体
func NewDatabaseCluster(
	orgName string,
	projectSlug string,
	title string,
	host string,
	port int,
	username string,
	password string,
) (*DatabaseCluster, error) {
	now := time.Now()
	var passwordObj *Password

	passwordObj, err := NewByPlain(password)
	if err != nil {
		return nil, err
	}

	dc := &DatabaseCluster{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Title:       title,
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    *passwordObj,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	dc.SetDefaults()
	if err := dc.Validate(); err != nil {
		return nil, err
	}
	return dc, nil
}
