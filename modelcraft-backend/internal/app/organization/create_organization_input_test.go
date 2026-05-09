package organization_test

import (
	"modelcraft/internal/app/organization"
	"testing"
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
