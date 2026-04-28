package rbac

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRowPolicy_Validate(t *testing.T) {
	t.Run("select denied cascades", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: false},
			Insert: InsertPolicy{Allowed: true, Scope: ScopeAll},
			Update: UpdatePolicy{Allowed: false},
			Delete: DeletePolicy{Allowed: false},
		}
		assert.Error(t, policy.Validate())
	})

	t.Run("custom select requires predicate", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeCustom},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{Allowed: false},
			Delete: DeletePolicy{Allowed: false},
		}
		assert.Error(t, policy.Validate())
	})

	t.Run("scope all ignores predicate", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeAll, Predicate: rawJSON(`{"id":{"_eq":"1"}}`)},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{Allowed: false},
			Delete: DeletePolicy{Allowed: false},
		}
		assert.NoError(t, policy.Validate())
	})

	t.Run("update custom check requires check", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeAll},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{Allowed: true, Scope: ScopeAll, CheckScope: ScopeCustom},
			Delete: DeletePolicy{Allowed: false},
		}
		assert.Error(t, policy.Validate())
	})

	t.Run("update disabled ignores extra fields", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeAll},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{
				Allowed:    false,
				Scope:      ScopeCustom,
				Predicate:  rawJSON(`{"owner":{"_eq":"$endUserId"}}`),
				CheckScope: ScopeCustom,
				Check:      rawJSON(`{"owner":{"_eq":"$endUserId"}}`),
			},
			Delete: DeletePolicy{Allowed: false},
		}
		assert.NoError(t, policy.Validate())
	})
}

func TestRowPolicy_Normalization(t *testing.T) {
	t.Run("scope all clears predicate", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeAll, Predicate: rawJSON(`{"id":{"_eq":"1"}}`)},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{Allowed: false},
			Delete: DeletePolicy{Allowed: false},
		}
		policy.Normalize()
		assert.Nil(t, policy.Select.Predicate)
	})

	t.Run("allowed false clears fields", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: false, Scope: ScopeCustom, Predicate: rawJSON(`{"id":{"_eq":"1"}}`)},
			Insert: InsertPolicy{Allowed: false, Scope: ScopeCustom, Check: rawJSON(`{"id":{"_eq":"1"}}`)},
			Update: UpdatePolicy{
				Allowed:    false,
				Scope:      ScopeCustom,
				CheckScope: ScopeCustom,
				Check:      rawJSON(`{"id":{"_eq":"1"}}`),
			},
			Delete: DeletePolicy{Allowed: false, Scope: ScopeCustom, Predicate: rawJSON(`{"id":{"_eq":"1"}}`)},
		}
		policy.Normalize()
		assert.Equal(t, PolicyScope(""), policy.Select.Scope)
		assert.Nil(t, policy.Select.Predicate)
		assert.Equal(t, PolicyScope(""), policy.Update.CheckScope)
	})

	t.Run("update check scope default all", func(t *testing.T) {
		policy := &RowPolicy{
			Select: SelectPolicy{Allowed: true, Scope: ScopeAll},
			Insert: InsertPolicy{Allowed: false},
			Update: UpdatePolicy{Allowed: true, Scope: ScopeAll},
			Delete: DeletePolicy{Allowed: false},
		}
		policy.Normalize()
		assert.Equal(t, ScopeAll, policy.Update.CheckScope)
	})
}

func TestRowPolicy_HiddenRule(t *testing.T) {
	cases := []struct {
		name   string
		insert bool
		update bool
		delete bool
		hasErr bool
	}{
		{name: "violate with insert", insert: true, hasErr: true},
		{name: "all denied pass", hasErr: false},
		{name: "violate with update", update: true, hasErr: true},
		{name: "violate with delete", delete: true, hasErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy := &RowPolicy{
				Select: SelectPolicy{Allowed: false},
				Insert: InsertPolicy{Allowed: tc.insert, Scope: ScopeAll},
				Update: UpdatePolicy{Allowed: tc.update, Scope: ScopeAll},
				Delete: DeletePolicy{Allowed: tc.delete, Scope: ScopeAll},
			}
			err := policy.Validate()
			if tc.hasErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestColumnPolicy_Writable(t *testing.T) {
	t.Run("default writable true", func(t *testing.T) {
		rule := ColumnRule{FieldName: "status", Mode: ColumnAccessModeVisible}
		assert.True(t, rule.IsWritable())
	})

	t.Run("hidden writable true is allowed", func(t *testing.T) {
		w := true
		rule := ColumnRule{FieldName: "id_card", Mode: ColumnAccessModeHidden, Writable: &w}
		assert.True(t, rule.IsWritable())
	})

	t.Run("masked writable true is allowed", func(t *testing.T) {
		w := true
		rule := ColumnRule{FieldName: "phone", Mode: ColumnAccessModeMasked, Writable: &w}
		assert.True(t, rule.IsWritable())
	})

	t.Run("visible writable false", func(t *testing.T) {
		w := false
		rule := ColumnRule{FieldName: "status", Mode: ColumnAccessModeVisible, Writable: &w}
		assert.False(t, rule.IsWritable())
	})
}

func rawJSON(v string) json.RawMessage {
	return json.RawMessage(v)
}
