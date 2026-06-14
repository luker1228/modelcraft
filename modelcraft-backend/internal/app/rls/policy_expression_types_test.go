package rls

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionModeAllowedRoot(t *testing.T) {
	require.True(t, PolicyExpressionModeUsing.AllowsRoot("row"))
	require.True(t, PolicyExpressionModeUsing.AllowsRoot("auth"))
	require.False(t, PolicyExpressionModeUsing.AllowsRoot("input"))

	require.True(t, PolicyExpressionModeCheck.AllowsRoot("input"))
	require.True(t, PolicyExpressionModeCheck.AllowsRoot("auth"))
	require.False(t, PolicyExpressionModeCheck.AllowsRoot("row"))
}

func TestIsLegacyJSONExpression(t *testing.T) {
	require.True(t, IsLegacyJSONExpression(`{"owner_id":{"equals":"{{user_id}}"}}`))
	require.True(t, IsLegacyJSONExpression(`true`))
	require.False(t, IsLegacyJSONExpression(`row.owner_id == auth.userid`))
	require.False(t, IsLegacyJSONExpression(` input.owner_id == auth.userid `))
}
