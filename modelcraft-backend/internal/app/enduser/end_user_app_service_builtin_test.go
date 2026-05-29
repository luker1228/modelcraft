package enduser_test

import (
	"context"
	"testing"
	"time"

	appenduser "modelcraft/internal/app/enduser"
	domainenduser "modelcraft/internal/domain/enduser"
)

// fakeEndUserRepo implements domainenduser.EndUserRepository for testing.
type fakeEndUserRepo struct {
	users map[string]*domainenduser.EndUser
}

func (f *fakeEndUserRepo) Save(_ context.Context, u *domainenduser.EndUser) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeEndUserRepo) GetByID(_ context.Context, _, id string) (*domainenduser.EndUser, error) {
	return f.users[id], nil
}

func (f *fakeEndUserRepo) GetByUsername(_ context.Context, _, username string) (*domainenduser.EndUser, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeEndUserRepo) GetByUsernameGlobal(_ context.Context, username string) (*domainenduser.EndUser, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeEndUserRepo) GetByPhone(_ context.Context, _, _ string) (*domainenduser.EndUser, error) {
	return nil, nil //nolint:nilnil
}

func (f *fakeEndUserRepo) UpdateStatus(_ context.Context, _, id string, isForbidden bool) error {
	if u, ok := f.users[id]; ok {
		u.IsForbidden = isForbidden
	}
	return nil
}

func (f *fakeEndUserRepo) Delete(_ context.Context, _, id string) error {
	delete(f.users, id)
	return nil
}

func (f *fakeEndUserRepo) ListWithTotal(
	_ context.Context, _ domainenduser.ListEndUsersQuery,
) ([]*domainenduser.EndUser, int64, error) {
	return nil, 0, nil
}

func (f *fakeEndUserRepo) ListAccessibleProjectsByRoleAssignment(
	_ context.Context, _, _ string,
) ([]domainenduser.AccessibleProject, error) {
	return nil, nil
}

func (f *fakeEndUserRepo) HasProjectAccessByRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func (f *fakeEndUserRepo) ListAllProjectsByOrg(_ context.Context, _ string) ([]domainenduser.AccessibleProject, error) {
	return []domainenduser.AccessibleProject{}, nil
}

func (f *fakeEndUserRepo) UpdatePassword(_ context.Context, _, _ string, _ domainenduser.HashedPassword) error {
	return nil
}

func makeNormalUser() *domainenduser.EndUser {
	pwd, _ := domainenduser.NewHashedPasswordFromPlain("Password1")
	u, _ := domainenduser.NewEndUser("user-id", "myorg", "normaluser", pwd)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return u
}

func TestDeleteEndUser_Allowed(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"user-id": makeNormalUser(),
	}}
	svc := appenduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.DeleteEndUserDirect(context.Background(), "myorg", "user-id")
	if err != nil {
		t.Fatalf("expected no error when deleting normal user, got: %v", err)
	}
}

func TestUpdateEndUserStatus_Allowed(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"user-id": makeNormalUser(),
	}}
	svc := appenduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.UpdateEndUserStatusDirect(context.Background(), "myorg", "user-id", true)
	if err != nil {
		t.Fatalf("expected no error when disabling normal user, got: %v", err)
	}
}
