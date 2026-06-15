package modelruntime_test

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/modelruntime"
	"testing"
)

func TestFindEndUserRefFieldName_Found(t *testing.T) {
	fields := map[string]*modelruntime.RuntimeField{
		"owner_user": {
			Name: "owner_user",
			Type: &modeldesign.FieldType{Format: modeldesign.FormatEndUserRef},
		},
		"title": {
			Name: "title",
			Type: &modeldesign.FieldType{Format: modeldesign.FormatString},
		},
	}
	got := modelruntime.FindEndUserRefFieldName(fields)
	if got != "owner_user" {
		t.Errorf("expected 'owner_user', got %q", got)
	}
}

func TestFindEndUserRefFieldName_NotFound(t *testing.T) {
	fields := map[string]*modelruntime.RuntimeField{
		"title": {Name: "title", Type: &modeldesign.FieldType{Format: modeldesign.FormatString}},
	}
	got := modelruntime.FindEndUserRefFieldName(fields)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestBuildRowFilter_Allowed(t *testing.T) {
	// Policy without UsingExpr → default SELF scope → owner filter injected
	perms := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect},
		},
	}
	filter := modelruntime.BuildRowFilter(perms, modelruntime.ActionSelect, "owner_user", "user-abc")
	if filter == nil {
		t.Fatal("expected non-nil filter (default SELF scope when no UsingExpr)")
	}
	if filter["owner_user"] != "user-abc" {
		t.Errorf("expected owner_user=user-abc, got %v", filter["owner_user"])
	}
}

func TestBuildRowFilter_AllScopeWithUsing(t *testing.T) {
	// Policy WITH UsingExpr → RLS handles scoping → no owner filter injected
	perms := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect, UsingExpr: "true"},
		},
	}
	if got := modelruntime.BuildRowFilter(perms, modelruntime.ActionSelect, "owner_user", "user-abc"); got != nil {
		t.Error("policy with UsingExpr should NOT inject owner filter (RLS handles scoping)")
	}
}

func TestBuildRowFilter_NotAllowed(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionInsert},
		},
	}
	if got := modelruntime.BuildRowFilter(perms, modelruntime.ActionSelect, "owner_user", "user-abc"); got != nil {
		t.Error("not-allowed should return nil filter")
	}
}

func TestBuildRowFilter_NilPerms(t *testing.T) {
	if got := modelruntime.BuildRowFilter(nil, modelruntime.ActionSelect, "owner_user", "user-abc"); got != nil {
		t.Error("nil perms (tenant admin) should return nil filter")
	}
}

func TestBuildRowFilter_NoOwnerField(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect},
		},
	}
	if got := modelruntime.BuildRowFilter(perms, modelruntime.ActionSelect, "", "user-abc"); got != nil {
		t.Error("empty ownerField should return nil filter")
	}
}

func TestActionGate_DeniesWhenAllFalse(t *testing.T) {
	perms := &modelruntime.ResolvedModelPermissions{} // all false
	for _, action := range []modelruntime.Action{
		modelruntime.ActionSelect, modelruntime.ActionInsert,
		modelruntime.ActionUpdate, modelruntime.ActionDelete,
	} {
		if err := perms.CheckAction(action); err == nil {
			t.Errorf("action %s should be denied", action)
		}
	}
}
