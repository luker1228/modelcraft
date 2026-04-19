package modeldesign

import (
	"time"

	"modelcraft/internal/domain/rls"
)

// ModelRLSPolicy RLS 策略实体（五件套 JsonExpr）
type ModelRLSPolicy struct {
	ModelID         string       `json:"modelId"`
	SelectPredicate rls.JsonExpr `json:"selectPredicate"` // SELECT USING
	InsertCheck     rls.JsonExpr `json:"insertCheck"`     // INSERT WITH CHECK
	UpdatePredicate rls.JsonExpr `json:"updatePredicate"` // UPDATE USING
	UpdateCheck     rls.JsonExpr `json:"updateCheck"`     // UPDATE WITH CHECK
	DeletePredicate rls.JsonExpr `json:"deletePredicate"` // DELETE USING
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

// GetPreset 返回当前策略匹配的 Preset，自定义组合返回 nil
func (p *ModelRLSPolicy) GetPreset() *rls.RLSPreset {
	return p.matchPreset()
}

// matchPreset 匹配预设策略
func (p *ModelRLSPolicy) matchPreset() *rls.RLSPreset {
	// 五件套全为 OWNER_EQUALS_USER → READ_WRITE_OWNER
	if p.SelectPredicate.IsOwnerEqualsUser() &&
		p.InsertCheck.IsOwnerEqualsUser() &&
		p.UpdatePredicate.IsOwnerEqualsUser() &&
		p.UpdateCheck.IsOwnerEqualsUser() &&
		p.DeletePredicate.IsOwnerEqualsUser() {
		preset := rls.RLSPresetReadWriteOwner
		return &preset
	}

	// select=true, 其余=OWNER_EQUALS_USER → READ_ALL_WRITE_OWNER
	if p.SelectPredicate.IsTrue() &&
		p.InsertCheck.IsOwnerEqualsUser() &&
		p.UpdatePredicate.IsOwnerEqualsUser() &&
		p.UpdateCheck.IsOwnerEqualsUser() &&
		p.DeletePredicate.IsOwnerEqualsUser() {
		preset := rls.RLSPresetReadAllWriteOwner
		return &preset
	}

	// select=true, 其余=false → READ_ALL
	if p.SelectPredicate.IsTrue() &&
		p.InsertCheck.IsFalse() &&
		p.UpdatePredicate.IsFalse() &&
		p.UpdateCheck.IsFalse() &&
		p.DeletePredicate.IsFalse() {
		preset := rls.RLSPresetReadAll
		return &preset
	}

	// 五件套全 true → READ_WRITE_ALL
	if p.SelectPredicate.IsTrue() &&
		p.InsertCheck.IsTrue() &&
		p.UpdatePredicate.IsTrue() &&
		p.UpdateCheck.IsTrue() &&
		p.DeletePredicate.IsTrue() {
		preset := rls.RLSPresetReadWriteAll
		return &preset
	}

	// 五件套全 false → NO_ACCESS
	if p.SelectPredicate.IsFalse() &&
		p.InsertCheck.IsFalse() &&
		p.UpdatePredicate.IsFalse() &&
		p.UpdateCheck.IsFalse() &&
		p.DeletePredicate.IsFalse() {
		preset := rls.RLSPresetNoAccess
		return &preset
	}

	// 其他组合 → nil (自定义)
	return nil
}

// ApplyPreset 应用预设策略
func (p *ModelRLSPolicy) ApplyPreset(preset rls.RLSPreset) {
	switch preset {
	case rls.RLSPresetReadWriteOwner:
		ownerEqualsUser := rls.JsonExpr(`{"owner":{"_eq":{"_auth":"uid"}}}`)
		p.SelectPredicate = ownerEqualsUser
		p.InsertCheck = ownerEqualsUser
		p.UpdatePredicate = ownerEqualsUser
		p.UpdateCheck = ownerEqualsUser
		p.DeletePredicate = ownerEqualsUser
	case rls.RLSPresetReadAllWriteOwner:
		ownerEqualsUser := rls.JsonExpr(`{"owner":{"_eq":{"_auth":"uid"}}}`)
		p.SelectPredicate = rls.JsonExpr(`true`)
		p.InsertCheck = ownerEqualsUser
		p.UpdatePredicate = ownerEqualsUser
		p.UpdateCheck = ownerEqualsUser
		p.DeletePredicate = ownerEqualsUser
	case rls.RLSPresetReadAll:
		p.SelectPredicate = rls.JsonExpr(`true`)
		p.InsertCheck = rls.JsonExpr(`false`)
		p.UpdatePredicate = rls.JsonExpr(`false`)
		p.UpdateCheck = rls.JsonExpr(`false`)
		p.DeletePredicate = rls.JsonExpr(`false`)
	case rls.RLSPresetReadWriteAll:
		allTrue := rls.JsonExpr(`true`)
		p.SelectPredicate = allTrue
		p.InsertCheck = allTrue
		p.UpdatePredicate = allTrue
		p.UpdateCheck = allTrue
		p.DeletePredicate = allTrue
	case rls.RLSPresetNoAccess:
		allFalse := rls.JsonExpr(`false`)
		p.SelectPredicate = allFalse
		p.InsertCheck = allFalse
		p.UpdatePredicate = allFalse
		p.UpdateCheck = allFalse
		p.DeletePredicate = allFalse
	}
}

// IsUsingTrue 判断 USING 谓词是否为 true（全量访问）
func (p *ModelRLSPolicy) IsUsingTrue() bool {
	return p.SelectPredicate.IsTrue() &&
		p.UpdatePredicate.IsTrue() &&
		p.DeletePredicate.IsTrue()
}

// IsDenyAll 判断是否为 DENY ALL 策略
func (p *ModelRLSPolicy) IsDenyAll() bool {
	return p.SelectPredicate.IsFalse() &&
		p.InsertCheck.IsFalse() &&
		p.UpdatePredicate.IsFalse() &&
		p.UpdateCheck.IsFalse() &&
		p.DeletePredicate.IsFalse()
}
