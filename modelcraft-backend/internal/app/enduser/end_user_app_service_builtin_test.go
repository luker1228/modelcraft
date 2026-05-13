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

func (f *fakeEndUserRepo) GetBuiltinByOrg(_ context.Context, _ string) (*domainenduser.EndUser, error) {
	for _, u := range f.users {
		if u.IsBuiltin {
			return u, nil
		}
	}
	return nil, nil
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

func (f *fakeEndUserRepo) ListWithTotal(_ context.Context, _ domainenduser.ListEndUsersQuery) ([]*domainenduser.EndUser, int64, error) {
	return nil, 0, nil
}

func (f *fakeEndUserRepo) ListAccessibleProjectsByRoleAssignment(_ context.Context, _, _ string) ([]domainenduser.AccessibleProject, error) {
	return nil, nil
}

func (f *fakeEndUserRepo) HasProjectAccessByRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func (f *fakeEndUserRepo) UpdatePassword(_ context.Context, _, _ string, _ domainenduser.HashedPassword) error {
	return nil
}

func makeBuiltinUser() *domainenduser.EndUser {
	pwd, _ := domainenduser.NewHashedPasswordFromPlain("Password1")
	u, _ := domainenduser.NewBuiltinEndUser("builtin-id", "myorg", "creator", pwd)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return u
}

func TestDeleteBuiltinEndUser_ReturnsError(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"builtin-id": makeBuiltinUser(),
	}}
	svc := appenduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.DeleteEndUserDirect(context.Background(), "myorg", "builtin-id")
	if err == nil {
		t.Fatal("expected error when deleting builtin user, got nil")
	}
}

func TestDisableBuiltinEndUser_ReturnsError(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"builtin-id": makeBuiltinUser(),
	}}
	svc := appenduser.NewEndUserManagementAppServiceWithRepo(repo)

	err := svc.UpdateEndUserStatusDirect(context.Background(), "myorg", "builtin-id", true)
	if err == nil {
		t.Fatal("expected error when disabling builtin user, got nil")
	}
}

func TestEnableBuiltinEndUser_Allowed(t *testing.T) {
	repo := &fakeEndUserRepo{users: map[string]*domainenduser.EndUser{
		"builtin-id": makeBuiltinUser(),
	}}
	svc := appenduser.NewEndUserManagementAppServiceWithRepo(repo)

	// Enabling (isForbidden=false) should be allowed even for builtin
	err := svc.UpdateEndUserStatusDirect(context.Background(), "myorg", "builtin-id", false)
	if err != nil {
		t.Fatalf("expected no error when enabling builtin user, got: %v", err)
	}
}
