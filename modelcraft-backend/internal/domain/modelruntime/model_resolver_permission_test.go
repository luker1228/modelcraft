package modelruntime_test

import (
	"modelcraft/internal/domain/modelruntime"
	"testing"
)

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
