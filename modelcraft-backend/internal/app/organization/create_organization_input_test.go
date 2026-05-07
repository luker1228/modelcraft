package organization_test

import (
	"testing"

	"modelcraft/internal/app/organization"
)

func TestCreateOrganizationInput_HasEndUserAdminPassword(t *testing.T) {
	input := organization.CreateOrganizationInput{
		DisplayName:          "Test Org",
		OwnerUserID:          "user-1",
		EndUserAdminPassword: "Password1",
	}
	if input.EndUserAdminPassword == "" {
		t.Error("expected EndUserAdminPassword to be set")
	}
}
