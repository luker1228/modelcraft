package modeldesign

import (
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// LogicalFKDirection 逻辑外键方向枚举
type LogicalFKDirection string

const (
	DirectionNormal  LogicalFKDirection = "normal"
	DirectionReverse LogicalFKDirection = "reverse"
)

// LogicalForeignKey 逻辑外键领域实体
// 每个 FK 关系由两条记录组成，共享同一个 pair_id：
//   - direction=normal: 拥有 FK 列的模型（source_fields 是 FK 列）
//   - direction=reverse: 被引用的模型（镜像存储，source_fields 是被引用列）
type LogicalForeignKey struct {
	ID           string             `json:"id"`
	PairID       string             `json:"pairId"`
	Direction    LogicalFKDirection `json:"direction"`
	ModelID      string             `json:"modelId"`
	ModelName    string             `json:"modelName"`
	RefModelID   string             `json:"refModelId"`
	RefModelName string             `json:"refModelName"`
	SourceFields []string           `json:"sourceFields"`
	TargetFields []string           `json:"targetFields"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// Validate 验证逻辑外键实体的有效性
func (lf *LogicalForeignKey) Validate() error {
	if lf.ID == "" {
		return bizerrors.New("logical foreign key ID cannot be empty")
	}
	if lf.PairID == "" {
		return bizerrors.New("logical foreign key PairID cannot be empty")
	}
	if lf.Direction != DirectionNormal && lf.Direction != DirectionReverse {
		return bizerrors.Errorf("logical foreign key direction must be '%s' or '%s', got '%s'",
			DirectionNormal, DirectionReverse, lf.Direction)
	}
	if lf.ModelID == "" {
		return bizerrors.New("logical foreign key ModelID cannot be empty")
	}
	if lf.ModelName == "" {
		return bizerrors.New("logical foreign key ModelName cannot be empty")
	}
	if lf.RefModelID == "" {
		return bizerrors.New("logical foreign key RefModelID cannot be empty")
	}
	if lf.RefModelName == "" {
		return bizerrors.New("logical foreign key RefModelName cannot be empty")
	}
	if len(lf.SourceFields) == 0 {
		return bizerrors.New("logical foreign key SourceFields cannot be empty")
	}
	if len(lf.TargetFields) == 0 {
		return bizerrors.New("logical foreign key TargetFields cannot be empty")
	}
	if len(lf.SourceFields) != len(lf.TargetFields) {
		return bizerrors.Errorf(
			"logical foreign key SourceFields and TargetFields count mismatch: %d vs %d",
			len(lf.SourceFields), len(lf.TargetFields),
		)
	}
	return nil
}

// IsNormal 判断是否为 normal 方向（拥有 FK 列的一侧）
func (lf *LogicalForeignKey) IsNormal() bool {
	return lf.Direction == DirectionNormal
}

// IsReverse 判断是否为 reverse 方向（被引用的一侧）
func (lf *LogicalForeignKey) IsReverse() bool {
	return lf.Direction == DirectionReverse
}
