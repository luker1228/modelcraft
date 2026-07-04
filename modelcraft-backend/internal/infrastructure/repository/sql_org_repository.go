// Package repository provides sqlc-based implementations of domain repository interfaces.
package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/organization"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/domain/user"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"

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
		Phone:       row.Phone,
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
		OrgName:      row.OrgName,
		IsAdmin:      row.IsAdmin,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// userPhoneRowToDomain converts GetUserByPhoneInOrgRow to domain User.
func userPhoneRowToDomain(row dbgen.GetUserByPhoneInOrgRow) *user.User {
	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}
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
		OrgName:      row.OrgName,
		IsAdmin:      row.IsAdmin,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// userNameRowToDomain converts GetUserByNameInOrgRow to domain User.
func userNameRowToDomain(row dbgen.GetUserByNameInOrgRow) *user.User {
	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}
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
		OrgName:      row.OrgName,
		IsAdmin:      row.IsAdmin,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
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
		Phone:       org.Phone,
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

// GetByPhone retrieves an organization by its registered phone number (globally unique).
// Returns nil, shared.NewNotFoundError when no organization matches the given phone.
func (r *SqlOrganizationRepository) GetByPhone(ctx context.Context, phone string) (*organization.Organization, error) {
	row, err := r.q.GetOrganizationByPhone(ctx, phone)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("organization not found by phone: " + phone)
		}
		return nil, bizerrors.Wrapf(err, "failed to get organization by phone: %s", phone)
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

// ExistsByPhone checks whether an organization with the given phone already exists.
func (r *SqlOrganizationRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	count, err := r.q.ExistsOrganizationByPhone(ctx, phone)
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check organization phone existence: %s", phone)
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
		OrgName:     u.OrgName,
		IsAdmin:     u.IsAdmin,
		Status:      u.Status,
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

// GetByPhone retrieves a user by org and phone number.
// Returns nil, shared.NewNotFoundError when no user matches.
func (r *SqlUserRepository) GetByPhone(ctx context.Context, orgName, phone string) (*user.User, error) {
	row, err := r.q.GetUserByPhoneInOrg(ctx, dbgen.GetUserByPhoneInOrgParams{OrgName: orgName, Phone: phone})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by phone in org: " + orgName)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by phone in org: %s", orgName)
	}
	return userPhoneRowToDomain(row), nil
}

// ExistsByPhone checks whether a user with the given phone exists within the org.
func (r *SqlUserRepository) ExistsByPhone(ctx context.Context, orgName, phone string) (bool, error) {
	exists, err := r.q.ExistsByPhoneInOrg(ctx, dbgen.ExistsByPhoneInOrgParams{OrgName: orgName, Phone: phone})
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check phone existence in org: %s", orgName)
	}
	return exists, nil
}

// GetByName retrieves a user by org and username.
// Returns nil, shared.NewNotFoundError when no user matches.
func (r *SqlUserRepository) GetByName(ctx context.Context, orgName, name string) (*user.User, error) {
	row, err := r.q.GetUserByNameInOrg(ctx, dbgen.GetUserByNameInOrgParams{OrgName: orgName, Name: name})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by name in org: " + orgName)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by name in org: %s", orgName)
	}
	return userNameRowToDomain(row), nil
}

// ExistsByName checks whether a user with the given name exists within the org.
func (r *SqlUserRepository) ExistsByName(ctx context.Context, orgName, name string) (bool, error) {
	exists, err := r.q.ExistsByUserNameInOrg(ctx, dbgen.ExistsByUserNameInOrgParams{OrgName: orgName, Name: name})
	if err != nil {
		return false, bizerrors.Wrapf(err, "failed to check user name existence in org: %s", orgName)
	}
	return exists, nil
}

// GetByNameGlobal retrieves a user by username across all orgs (used for admin login).
func (r *SqlUserRepository) GetByNameGlobal(ctx context.Context, name string) (*user.User, error) {
	row, err := r.q.GetUserByNameGlobal(ctx, name)
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user not found by name: " + name)
		}
		return nil, bizerrors.Wrapf(err, "failed to get user by name globally")
	}
	var externalID string
	if row.ExternalID.Valid {
		externalID = row.ExternalID.String
	}
	var phone user.PhoneNumber
	if row.Phone != "" {
		if p, err := user.NewPhoneNumber(row.Phone); err == nil {
			phone = p
		}
	}
	return &user.User{
		ID:           row.ID,
		ExternalID:   externalID,
		Name:         row.Name,
		Phone:        phone,
		PasswordHash: row.PasswordHash,
		OrgName:      row.OrgName,
		IsAdmin:      row.IsAdmin,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

// ListByOrg returns all active users belonging to the given org.
func (r *SqlUserRepository) ListByOrg(ctx context.Context, orgName string) ([]*user.User, error) {
	rows, err := r.q.ListUsersByOrgWithName(ctx, orgName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list users by org: %s", orgName)
	}
	result := make([]*user.User, len(rows))
	for i, row := range rows {
		result[i] = &user.User{
			ID:      row.ID,
			Name:    row.Name,
			OrgName: row.OrgName,
			IsAdmin: row.IsAdmin,
			Status:  row.Status,
		}
	}
	return result, nil
}

// Compile-time interface satisfaction checks.
var (
	_ organization.OrganizationRepository = (*SqlOrganizationRepository)(nil)
	_ user.UserRepository                 = (*SqlUserRepository)(nil)
)
