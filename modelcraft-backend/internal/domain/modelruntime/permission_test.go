package modelruntime_test

import (
	"modelcraft/internal/domain/modelruntime"
	"testing"
)

func TestResolvedModelPermissions_CheckAction_NilSkipsAll(t *testing.T) {
	var p *modelruntime.ResolvedModelPermissions
	for _, action := range []modelruntime.Action{
		modelruntime.ActionSelect, modelruntime.ActionInsert,
		modelruntime.ActionUpdate, modelruntime.ActionDelete,
	} {
		if err := p.CheckAction(action); err != nil {
			t.Errorf("nil permissions should allow action %s, got error: %v", action, err)
		}
	}
}

func TestResolvedModelPermissions_CheckAction_DeniesWhenNotAllowed(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect},
		},
	}
	if err := p.CheckAction(modelruntime.ActionSelect); err != nil {
		t.Errorf("expected select to be allowed, got: %v", err)
	}
	if err := p.CheckAction(modelruntime.ActionInsert); err == nil {
		t.Error("expected insert to be denied")
	}
	if err := p.CheckAction(modelruntime.ActionUpdate); err == nil {
		t.Error("expected update to be denied")
	}
	if err := p.CheckAction(modelruntime.ActionDelete); err == nil {
		t.Error("expected delete to be denied")
	}
}

func TestResolvedModelPermissions_CheckAction_UnknownAction(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect},
		},
	}
	if err := p.CheckAction(modelruntime.Action("UNKNOWN")); err == nil {
		t.Error("unknown action should be denied on non-nil permissions")
	}
}

func TestResolvedModelPermissions_Get(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Policies: []modelruntime.ResolvedPolicy{
			{Action: modelruntime.ActionSelect, },
			{Action: modelruntime.ActionInsert},
		},
	}
	if got := p.Get(modelruntime.ActionSelect); !got.Allowed {
		t.Errorf("Select: want {true,true}, got %+v", got)
	}
	if got := p.Get(modelruntime.ActionInsert); !got.Allowed {
		t.Errorf("Insert: want {true,false}, got %+v", got)
	}
	if got := p.Get(modelruntime.Action("UNKNOWN")); got.Allowed {
		t.Error("unknown action should be denied")
	}
	// nil receiver + unknown action should be denied
	var nilP *modelruntime.ResolvedModelPermissions
	if got := nilP.Get(modelruntime.Action("UNKNOWN")); got.Allowed {
		t.Error("nil receiver + unknown action should be denied")
	}
	// nil receiver + known action → tenant admin → allowed
	if got := nilP.Get(modelruntime.ActionSelect); !got.Allowed {
		t.Error("nil receiver + known action should be allowed (tenant admin)")
	}
}
