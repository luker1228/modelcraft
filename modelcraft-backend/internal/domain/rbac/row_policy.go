package rbac

import (
	"bytes"
	"encoding/json"
	"modelcraft/pkg/bizerrors"
	"strings"
)

// PolicyScope 行策略作用域。
type PolicyScope string

const (
	ScopeAll    PolicyScope = "all"
	ScopeCustom PolicyScope = "custom"
)

var validPolicyScopes = map[PolicyScope]struct{}{
	ScopeAll:    {},
	ScopeCustom: {},
}

func (s PolicyScope) isValid() bool {
	_, ok := validPolicyScopes[s]
	return ok
}

// SelectPolicy 读取策略。
type SelectPolicy struct {
	Allowed   bool            `json:"allowed"`
	Scope     PolicyScope     `json:"scope,omitempty"`
	Predicate json.RawMessage `json:"predicate,omitempty"`
}

// InsertPolicy 创建策略。
type InsertPolicy struct {
	Allowed bool            `json:"allowed"`
	Scope   PolicyScope     `json:"scope,omitempty"`
	Check   json.RawMessage `json:"check,omitempty"`
}

// UpdatePolicy 更新策略。
type UpdatePolicy struct {
	Allowed    bool            `json:"allowed"`
	Scope      PolicyScope     `json:"scope,omitempty"`
	Predicate  json.RawMessage `json:"predicate,omitempty"`
	CheckScope PolicyScope     `json:"check_scope,omitempty"`
	Check      json.RawMessage `json:"check,omitempty"`
}

// DeletePolicy 删除策略。
type DeletePolicy struct {
	Allowed   bool            `json:"allowed"`
	Scope     PolicyScope     `json:"scope,omitempty"`
	Predicate json.RawMessage `json:"predicate,omitempty"`
}

// RowPolicy 四类动作统一行策略。
type RowPolicy struct {
	Select SelectPolicy `json:"select"`
	Insert InsertPolicy `json:"insert"`
	Update UpdatePolicy `json:"update"`
	Delete DeletePolicy `json:"delete"`
}

// Validate 校验 rowPolicy 语义。
func (p *RowPolicy) Validate() error {
	if p == nil {
		return bizerrors.NewValidationError("rbac.row_policy.required: rowPolicy is required")
	}

	if !p.Select.Allowed {
		if p.Insert.Allowed || p.Update.Allowed || p.Delete.Allowed {
			return bizerrors.NewValidationError(
				"rbac.row_policy.hidden_rule_violation: select.allowed=false requires insert/update/delete all false")
		}
	}

	if err := validateSelectPolicy(p.Select); err != nil {
		return err
	}
	if err := validateInsertPolicy(p.Insert); err != nil {
		return err
	}
	if err := validateUpdatePolicy(p.Update); err != nil {
		return err
	}
	if err := validateDeletePolicy(p.Delete); err != nil {
		return err
	}
	return nil
}

// Normalize 归一化 rowPolicy，消除无效冗余字段。
func (p *RowPolicy) Normalize() {
	if p == nil {
		return
	}

	normalizeSelectPolicy(&p.Select)
	normalizeInsertPolicy(&p.Insert)
	normalizeUpdatePolicy(&p.Update)
	normalizeDeletePolicy(&p.Delete)

	if !p.Select.Allowed {
		normalizeSelectDenyCascade(&p.Insert, &p.Update, &p.Delete)
	}
}

func validateSelectPolicy(p SelectPolicy) error {
	if !p.Allowed {
		return nil
	}
	if !p.Scope.isValid() {
		return bizerrors.NewValidationError("rbac.row_policy.invalid_scope: select.scope must be all or custom")
	}
	if p.Scope == ScopeCustom && isJSONEmpty(p.Predicate) {
		return bizerrors.NewValidationError(
			"rbac.row_policy.custom_predicate_required: " +
				"select.predicate is required when scope=custom",
		)
	}
	return nil
}

func validateInsertPolicy(p InsertPolicy) error {
	if !p.Allowed {
		return nil
	}
	if !p.Scope.isValid() {
		return bizerrors.NewValidationError(
			"rbac.row_policy.invalid_scope: insert.scope must be all or custom",
		)
	}
	if p.Scope == ScopeCustom && isJSONEmpty(p.Check) {
		return bizerrors.NewValidationError(
			"rbac.row_policy.custom_check_required: " +
				"insert.check is required when scope=custom",
		)
	}
	return nil
}

func validateUpdatePolicy(p UpdatePolicy) error {
	if !p.Allowed {
		return nil
	}
	if !p.Scope.isValid() {
		return bizerrors.NewValidationError(
			"rbac.row_policy.invalid_scope: update.scope must be all or custom",
		)
	}
	if p.Scope == ScopeCustom && isJSONEmpty(p.Predicate) {
		return bizerrors.NewValidationError(
			"rbac.row_policy.custom_predicate_required: " +
				"update.predicate is required when scope=custom",
		)
	}
	if p.CheckScope == "" {
		p.CheckScope = ScopeAll
	}
	if !p.CheckScope.isValid() {
		return bizerrors.NewValidationError(
			"rbac.row_policy.invalid_check_scope: " +
				"update.check_scope must be all or custom",
		)
	}
	if p.CheckScope == ScopeCustom && isJSONEmpty(p.Check) {
		return bizerrors.NewValidationError(
			"rbac.row_policy.custom_check_required: " +
				"update.check is required when check_scope=custom",
		)
	}
	return nil
}

func validateDeletePolicy(p DeletePolicy) error {
	if !p.Allowed {
		return nil
	}
	if !p.Scope.isValid() {
		return bizerrors.NewValidationError(
			"rbac.row_policy.invalid_scope: delete.scope must be all or custom",
		)
	}
	if p.Scope == ScopeCustom && isJSONEmpty(p.Predicate) {
		return bizerrors.NewValidationError(
			"rbac.row_policy.custom_predicate_required: " +
				"delete.predicate is required when scope=custom",
		)
	}
	return nil
}

func normalizeSelectPolicy(p *SelectPolicy) {
	if !p.Allowed {
		p.Scope = ""
		p.Predicate = nil
		return
	}
	if p.Scope == "" {
		p.Scope = ScopeAll
	}
	if p.Scope == ScopeAll {
		p.Predicate = nil
	}
}

func normalizeInsertPolicy(p *InsertPolicy) {
	if !p.Allowed {
		p.Scope = ""
		p.Check = nil
		return
	}
	if p.Scope == "" {
		p.Scope = ScopeAll
	}
	if p.Scope == ScopeAll {
		p.Check = nil
	}
}

func normalizeUpdatePolicy(p *UpdatePolicy) {
	if !p.Allowed {
		p.Scope = ""
		p.Predicate = nil
		p.CheckScope = ""
		p.Check = nil
		return
	}
	if p.Scope == "" {
		p.Scope = ScopeAll
	}
	if p.Scope == ScopeAll {
		p.Predicate = nil
	}
	if p.CheckScope == "" {
		p.CheckScope = ScopeAll
	}
	if p.CheckScope == ScopeAll {
		p.Check = nil
	}
}

func normalizeDeletePolicy(p *DeletePolicy) {
	if !p.Allowed {
		p.Scope = ""
		p.Predicate = nil
		return
	}
	if p.Scope == "" {
		p.Scope = ScopeAll
	}
	if p.Scope == ScopeAll {
		p.Predicate = nil
	}
}

func normalizeSelectDenyCascade(insert *InsertPolicy, update *UpdatePolicy, delete *DeletePolicy) {
	insert.Allowed = false
	insert.Scope = ""
	insert.Check = nil

	update.Allowed = false
	update.Scope = ""
	update.Predicate = nil
	update.CheckScope = ""
	update.Check = nil

	delete.Allowed = false
	delete.Scope = ""
	delete.Predicate = nil
}

func isJSONEmpty(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" || trimmed == "{}" || trimmed == "[]" {
		return true
	}
	return bytes.Equal(bytes.TrimSpace(raw), []byte(""))
}
