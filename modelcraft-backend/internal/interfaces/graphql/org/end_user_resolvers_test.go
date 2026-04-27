package orggraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"testing"

	appEnduser "modelcraft/internal/app/enduser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEndUser_ServiceNotInitialized(t *testing.T) {
	resolver := &mutationResolver{&Resolver{}}

	payload, err := resolver.CreateEndUser(context.Background(), generated.CreateEndUserInput{
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "end-user service not initialized", invalidInput.Message)
}

func TestCreateEndUser_MissingOrgContext(t *testing.T) {
	resolver := &mutationResolver{&Resolver{EndUserMgmtAppService: &appEnduser.EndUserManagementAppService{}}}

	payload, err := resolver.CreateEndUser(context.Background(), generated.CreateEndUserInput{
		Username: "alice",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "orgName not found in context", invalidInput.Message)
}

func TestCreateEndUser_ParamValidation(t *testing.T) {
	resolver := &mutationResolver{&Resolver{}}

	payload, err := resolver.CreateEndUser(context.Background(), generated.CreateEndUserInput{
		Username: "",
		Password: "Password123",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "username is required", invalidInput.Message)
}
