package modelruntime_test

import (
	"testing"

	"modelcraft/internal/domain/modelruntime"
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
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: false},
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

func TestResolvedModelPermissions_Get(t *testing.T) {
	p := &modelruntime.ResolvedModelPermissions{
		Select: modelruntime.ActionPermission{Allowed: true, IsSelf: true},
		Insert: modelruntime.ActionPermission{Allowed: true, IsSelf: false},
		Update: modelruntime.ActionPermission{Allowed: false},
		Delete: modelruntime.ActionPermission{Allowed: false},
	}
	if got := p.Get(modelruntime.ActionSelect); !got.Allowed || !got.IsSelf {
		t.Errorf("Select: want {true,true}, got %+v", got)
	}
	if got := p.Get(modelruntime.ActionInsert); !got.Allowed || got.IsSelf {
		t.Errorf("Insert: want {true,false}, got %+v", got)
	}
	if got := p.Get(modelruntime.Action("UNKNOWN")); got.Allowed {
		t.Error("unknown action should be denied")
	}
}
