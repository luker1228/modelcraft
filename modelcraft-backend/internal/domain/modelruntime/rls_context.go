package modelruntime

import "modelcraft/internal/domain/rls"

// RLSContext captures all request-scoped runtime identity and policy state.
type RLSContext struct {
	Action         Action
	EndUserID      string
	IsAdmin        bool
	UserContext    *rls.UserContext
	Permissions    *ResolvedModelPermissions
	Snapshot       *RLSPolicySnapshot
	FastFail       bool
	FastFailReason string
}
