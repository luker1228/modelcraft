package dtos

import (
	"modelcraft/internal/domain/cluster"
)

// ConnectionInfo HTTP DTO层的连接信息
type ConnectionInfo struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"` // 始终使用明文密码，用于HTTP接口传输
	ConnectionTimeout *int   `json:"connectionTimeout,omitempty"`
}

// ToDomainEntity 转换为领域实体的ConnectionInfo
func (dto *ConnectionInfo) ToDomainEntity() (*cluster.ConnectionInfo, error) {
	// HTTP DTO中的密码是明文，创建明文Password实例
	passwdobj, err := cluster.NewByPlain(dto.Password)
	if err != nil {
		return nil, err
	}

	connectionTimeout := 0
	if dto.ConnectionTimeout != nil {
		connectionTimeout = *dto.ConnectionTimeout
	}

	return &cluster.ConnectionInfo{
		Host:              dto.Host,
		Port:              dto.Port,
		Username:          dto.Username,
		Password:          *passwdobj,
		ConnectionTimeout: connectionTimeout,
	}, nil
}

// FromDomainEntity 从领域实体转换为DTO
func (dto *ConnectionInfo) FromDomainEntity(entity cluster.ConnectionInfo) error {
	dto.Host = entity.Host
	dto.Port = entity.Port
	dto.Username = entity.Username

	// 获取明文密码用于返回（注意：实际应用中可能需要考虑安全性）
	plainPassword, err := entity.Password.GetPlainPassword()
	if err != nil {
		return err
	}
	dto.Password = plainPassword

	if entity.ConnectionTimeout != 0 {
		ct := entity.ConnectionTimeout
		dto.ConnectionTimeout = &ct
	}

	return nil
}
