// Package repository provides sqlc-based implementations of domain repository interfaces.
package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/membership"
	"modelcraft/internal/domain/organization"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/domain/user"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// OrganizationToDomain converts a dbgen.Organization row to a domain Organization entity.
func OrganizationToDomain(row dbgen.Organization) *organization.Organization {
	var displayName string
	if row.DisplayName.Valid {
		displayName = row.DisplayName.String
	}

	var ownerID string
	if row.OwnerID.Valid {
		ownerID = row.OwnerID.String
	}

	return &organization.Organization{
		Name:        row.Name,
		DisplayName: displayName,
		OwnerID:     ownerID,
		Status:      organization.OrgStatus(row.Status),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// UserToDomain converts a dbgen.User row to a domain User entity.
func UserToDomain(row dbgen.User) *user.User {
	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}

	// Reconstruct PhoneNumber value object from stored string.
	// DB may store empty string for OAuth users, so we tolerate construction failure.
	var phone user.PhoneNumber
	if row.Phone != "" {
		p, err := user.NewPhoneNumber(row.Phone)
		if err == nil {
			phone = p
		}
	}

	return &user.User{
		ID:           row.ID,
		ExternalID:   externalID,
		Name:         row.Name,
		Phone:        phone,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// MembershipToDomain converts a dbgen.UserOrganization row to a domain Membership entity.
func MembershipToDomain(row dbgen.UserOrganization) *membership.Membership {
	var invitedBy string
	if row.InvitedBy.Valid {
		invitedBy = row.InvitedBy.String
	}

	var invitedAt *time.Time
	if row.InvitedAt.Valid {
		t := row.InvitedAt.Time
		invitedAt = &t
	}

	var joinedAt *time.Time
	if row.JoinedAt.Valid {
		t := row.JoinedAt.Time
		joinedAt = &t
	}

	return &membership.Membership{
		ID:        row.ID,
		UserID:    row.UserID,
		OrgName:   row.OrgName,
		Status:    membership.MembershipStatus(row.Status),
		InvitedBy: invitedBy,
		InvitedAt: invitedAt,
		JoinedAt:  joinedAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

// StrToNullStr converts a plain string to sql.NullString.
// An empty string produces an invalid (NULL) NullString.
func StrToNullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// SqlOrganizationRepository is the sqlc-based implementation of organization.OrganizationRepository.
type SqlOrganizationRepository struct {
	q dbgen.Querier
}

// NewSqlOrganizationRepository creates a new SqlOrganizationRepository backed by the given sqlc Querier.
func NewSqlOrganizationRepository(q dbgen.Querier) organization.OrganizationRepository {
	return &SqlOrganizationRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create persists a new organization to the database.
func (r *SqlOrganizationRepository) Create(ctx context.Context, org *organization.Organization) error {
	params := dbgen.CreateOrganizationParams{
		Name:        org.Name,
		DisplayName: StrToNullStr(org.DisplayName),
		OwnerID:     StrToNullStr(org.OwnerID),
		Status:      string(org.Status),
	}

	return r.q.CreateOrganization(ctx, params)
}

// GetByName retrieves an organization by its unique name.
// Returns nil, shared.NewNotFoundError when no organization matches the given name.
func (r *SqlOrganizationRepository) GetByName(ctx context.Context, name string) (*organization.Organization, error) {
	row, err := r.q.GetOrganizationByName(ctx, name)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("organization not found by name: " + name)
		}
		return nil, bizerrors.Wrapf(err, "failed to get organization by name: %s", name)
	}

	return OrganizationToDomain(row), nil
}

// ListByUser returns all active organizations the given user belongs to.
func (r *SqlOrganizationRepository) ListByUser(
	ctx context.Context, userID string,
) ([]*organization.Organization, error) {
	rows, err := r.q.ListOrganizationsByUser(ctx, userID)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list organizations by user: %s", userID)
	}

	orgs := make([]*organization.Organization, len(rows))
	for i, row := range rows {
		orgs[i] = OrganizationToDomain(row)
	}

	return orgs, nil
}

// Update persists changes to an existing organization.
// Only display_name and status are updated; name is the primary key and is immutable via SQL.
func (r *SqlOrganizationRepository) Update(ctx context.Context, org *organization.Organization) error {
	params := dbgen.UpdateOrganizationParams{
		DisplayName: StrToNullStr(org.DisplayName),
		Status:      string(org.Status),
		Name:        org.Name,
	}

	return r.q.UpdateOrganization(ctx, params)
}

// ExistsByName checks whether an organization with the given name already exists.
func (r *SqlOrganizationRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.q.ExistsOrganizationByName(ctx, name)
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check organization name existence: %s", name)
	}

	return count > 0, nil
}

// SqlUserRepository is the sqlc-based implementation of user.UserRepository.
type SqlUserRepository struct {
	q dbgen.Querier
}

// NewSqlUserRepository creates a new SqlUserRepository backed by the given sqlc Querier.
func NewSqlUserRepository(q dbgen.Querier) user.UserRepository {
	return &SqlUserRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create persists a new user to the database.
func (r *SqlUserRepository) Create(ctx context.Context, u *user.User) error {
	params := dbgen.CreateUserParams{
		ID:           u.ID,
		ExternalID:   StrToNullStr(u.ExternalID),
		Name:         u.Name,
		Phone:        u.Phone.String(),
		PasswordHash: u.PasswordHash,
		// DisplayName is not part of the domain User entity; stored as NULL.
		DisplayName: sql.NullString{},
	}

	return r.q.CreateUser(ctx, params)
}

// GetByID retrieves a user by internal UUID.
// Returns nil, shared.NewNotFoundError when the user is not found.
func (r *SqlUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by id: " + id)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by id: %s", id)
	}

	return UserToDomain(row), nil
}

// GetByExternalID retrieves a user by the external authentication provider ID.
// Returns nil, shared.NewNotFoundError when no user matches the given externalID.
func (r *SqlUserRepository) GetByExternalID(ctx context.Context, externalID string) (*user.User, error) {
	row, err := r.q.GetUserByExternalID(ctx, StrToNullStr(externalID))
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by external id: " + externalID)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by external id: %s", externalID)
	}

	return UserToDomain(row), nil
}

// ExistsByExternalID checks whether a user with the given external ID already exists.
func (r *SqlUserRepository) ExistsByExternalID(ctx context.Context, externalID string) (bool, error) {
	count, err := r.q.ExistsUserByExternalID(ctx, StrToNullStr(externalID))
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check user external id existence: %s", externalID)
	}

	return count > 0, nil
}

// FindIDByExternalID retrieves the internal user ID by external authentication provider ID.
// Returns ("", false, nil) if no user matches the given externalID.
func (r *SqlUserRepository) FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error) {
	userID, err := r.q.FindIDByExternalID(ctx, StrToNullStr(externalID))
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return "", false, nil
		}
		return "", false, bizerrors.Wrapf(err, "failed to find user id by external id: %s", externalID)
	}

	return userID, true, nil
}

// GetByPhone retrieves a user by phone number.
// Returns nil, shared.NewNotFoundError when no user matches the given phone.
func (r *SqlUserRepository) GetByPhone(ctx context.Context, phone string) (*user.User, error) {
	row, err := r.q.GetUserByPhone(ctx, phone)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by phone: " + phone)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by phone: %s", phone)
	}

	// Convert GetUserByPhoneRow to domain User.
	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}

	var phoneVO user.PhoneNumber
	if row.Phone != "" {
		p, err := user.NewPhoneNumber(row.Phone)
		if err == nil {
			phoneVO = p
		}
	}

	return &user.User{
		ID:           row.ID,
		ExternalID:   externalID,
		Name:         row.Name,
		Phone:        phoneVO,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

// ExistsByPhone checks whether a user with the given phone number already exists.
func (r *SqlUserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	exists, err := r.q.ExistsByPhone(ctx, phone)
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check phone existence: %s", phone)
	}

	return exists, nil
}

// SqlMembershipRepository is the sqlc-based implementation of membership.MembershipRepository.
type SqlMembershipRepository struct {
	q dbgen.Querier
}

// NewSqlMembershipRepository creates a new SqlMembershipRepository backed by the given sqlc Querier.
func NewSqlMembershipRepository(q dbgen.Querier) membership.MembershipRepository {
	return &SqlMembershipRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create persists a new membership to the database.
func (r *SqlMembershipRepository) Create(ctx context.Context, m *membership.Membership) error {
	params := dbgen.CreateMembershipParams{
		ID:        m.ID,
		UserID:    m.UserID,
		OrgName:   m.OrgName,
		Status:    string(m.Status),
		InvitedBy: StrToNullStr(m.InvitedBy),
		InvitedAt: sqlerr.PtrToNullTime(m.InvitedAt),
		JoinedAt:  sqlerr.PtrToNullTime(m.JoinedAt),
	}

	return r.q.CreateMembership(ctx, params)
}

// GetByID retrieves a membership by its UUID.
// Returns nil, shared.NewNotFoundError when the membership is not found.
func (r *SqlMembershipRepository) GetByID(ctx context.Context, id string) (*membership.Membership, error) {
	row, err := r.q.GetMembershipByID(ctx, id)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("membership not found by id: " + id)
		}
		return nil, bizerrors.Wrapf(err, "failed to get membership by id: %s", id)
	}

	return MembershipToDomain(row), nil
}

// GetByUserAndOrg retrieves a membership by user ID and organization name.
// Returns ErrRecordNotFound when no matching membership exists.
func (r *SqlMembershipRepository) GetByUserAndOrg(
	ctx context.Context, userID, orgName string,
) (*membership.Membership, error) {
	row, err := r.q.GetMembershipByUserAndOrg(ctx, dbgen.GetMembershipByUserAndOrgParams{
		UserID:  userID,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("membership not found for user " + userID + " in org " + orgName)
		}
		return nil, bizerrors.Wrapf(err, "failed to get membership for user %s in org %s", userID, orgName)
	}

	return MembershipToDomain(row), nil
}

// ListByOrg returns all memberships for the given organization.
func (r *SqlMembershipRepository) ListByOrg(ctx context.Context, orgName string) ([]*membership.Membership, error) {
	rows, err := r.q.ListMembershipsByOrg(ctx, orgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list memberships by org: %s", orgName)
	}

	memberships := make([]*membership.Membership, len(rows))
	for i, row := range rows {
		memberships[i] = MembershipToDomain(row)
	}

	return memberships, nil
}

// ListByOrgWithUserName returns all memberships for the given organization, with user names joined
// from the users table via a LEFT JOIN.
func (r *SqlMembershipRepository) ListByOrgWithUserName(
	ctx context.Context, orgName string,
) ([]*membership.MembershipWithUserName, error) {
	rows, err := r.q.ListMembershipsWithUserName(ctx, orgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list memberships with user name for org: %s", orgName)
	}

	result := make([]*membership.MembershipWithUserName, len(rows))
	for i, row := range rows {
		// Reconstruct the canonical UserOrganization row from the joined result fields.
		uOrg := dbgen.UserOrganization{
			ID:        row.ID,
			UserID:    row.UserID,
			OrgName:   row.OrgName,
			Status:    row.Status,
			InvitedBy: row.InvitedBy,
			InvitedAt: row.InvitedAt,
			JoinedAt:  row.JoinedAt,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}

		result[i] = &membership.MembershipWithUserName{
			Membership: MembershipToDomain(uOrg),
			UserName:   row.UserName,
		}
	}

	return result, nil
}

// ListByUser returns all memberships for the given user.
func (r *SqlMembershipRepository) ListByUser(ctx context.Context, userID string) ([]*membership.Membership, error) {
	rows, err := r.q.ListMembershipsByUser(ctx, userID)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list memberships by user: %s", userID)
	}

	memberships := make([]*membership.Membership, len(rows))
	for i, row := range rows {
		memberships[i] = MembershipToDomain(row)
	}

	return memberships, nil
}

// CountByUser returns the total number of memberships (organizations) the user belongs to.
func (r *SqlMembershipRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	count, err := r.q.CountMembershipsByUser(ctx, userID)
	if err != nil {
		return 0, bizerrors.Wrapf(err, "failed to count memberships by user: %s", userID)
	}

	return count, nil
}

// ListByUserWithDetails returns active memberships for the given user with organization display
// names included. Results are ordered by joined_at descending and limited to limit entries.
// RoleName is intentionally left empty; it is managed via a separate user_roles table.
func (r *SqlMembershipRepository) ListByUserWithDetails(
	ctx context.Context,
	userID string,
	limit int,
) ([]*membership.MembershipWithDetails, error) {
	if limit <= 0 {
		return nil, bizerrors.Errorf("limit must be greater than 0, got %d", limit)
	}
	rows, err := r.q.ListMembershipsWithOrgDetails(ctx, dbgen.ListMembershipsWithOrgDetailsParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list memberships with details for user: %s", userID)
	}

	details := make([]*membership.MembershipWithDetails, len(rows))
	for i, row := range rows {
		var displayName string
		if row.OrgDisplayName.Valid {
			displayName = row.OrgDisplayName.String
		}

		var joinedAt time.Time
		if row.JoinedAt.Valid {
			joinedAt = row.JoinedAt.Time
		}

		details[i] = &membership.MembershipWithDetails{
			OrgName:     row.OrgName,
			DisplayName: displayName,
			RoleName:    "",
			JoinedAt:    joinedAt,
		}
	}

	return details, nil
}

// Update persists changes to an existing membership: status, invited_by, invited_at, joined_at.
func (r *SqlMembershipRepository) Update(ctx context.Context, m *membership.Membership) error {
	params := dbgen.UpdateMembershipParams{
		Status:    string(m.Status),
		InvitedBy: StrToNullStr(m.InvitedBy),
		InvitedAt: sqlerr.PtrToNullTime(m.InvitedAt),
		JoinedAt:  sqlerr.PtrToNullTime(m.JoinedAt),
		ID:        m.ID,
	}

	return r.q.UpdateMembership(ctx, params)
}

// Delete removes the membership identified by id from the database.
func (r *SqlMembershipRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteMembership(ctx, id)
}

// Compile-time interface satisfaction checks.
var (
	_ organization.OrganizationRepository = (*SqlOrganizationRepository)(nil)
	_ user.UserRepository                 = (*SqlUserRepository)(nil)
	_ membership.MembershipRepository     = (*SqlMembershipRepository)(nil)
)
