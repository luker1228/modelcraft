// Package rls provides RLS (Row Level Security) repository implementations.
package rls

import (
	"context"
	"encoding/json"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"

	domainRLS "modelcraft/internal/domain/rls"
)

// AuthSchemaRepository defines the interface for project auth schema persistence operations.
// This interface should be moved to the domain layer when fully implemented.
type AuthSchemaRepository interface {
	// GetByProjectID retrieves an auth schema by org name and project slug.
	// Returns nil, nil if the schema does not exist.
	GetByProjectID(ctx context.Context, orgName, projectSlug string) (*domainRLS.AuthSchema, error)

	// Save saves an auth schema (upsert operation).
	Save(ctx context.Context, orgName, projectSlug string, authSchema *domainRLS.AuthSchema) error

	// DeleteByProjectID deletes the auth schema for a given project.
	DeleteByProjectID(ctx context.Context, orgName, projectSlug string) error
}

// SqlAuthSchemaRepository is the sqlc-based implementation of AuthSchemaRepository.
type SqlAuthSchemaRepository struct {
	q dbgen.Querier
}

// NewSqlAuthSchemaRepository creates a new SqlAuthSchemaRepository.
func NewSqlAuthSchemaRepository(q dbgen.Querier) AuthSchemaRepository {
	return &SqlAuthSchemaRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// GetByProjectID retrieves an auth schema by org name and project slug.
// Returns nil, nil if the schema does not exist.
func (r *SqlAuthSchemaRepository) GetByProjectID(
	ctx context.Context,
	orgName, projectSlug string,
) (*domainRLS.AuthSchema, error) {
	row, err := r.q.GetProjectAuthSchema(ctx, dbgen.GetProjectAuthSchemaParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // nil result with nil error indicates not found
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	var variables []domainRLS.AuthVariable
	if err := json.Unmarshal(row.Variables, &variables); err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	return &domainRLS.AuthSchema{
		ProjectID: row.ProjectSlug,
		Variables: variables,
	}, nil
}

// Save saves an auth schema (upsert operation).
func (r *SqlAuthSchemaRepository) Save(
	ctx context.Context,
	orgName, projectSlug string,
	authSchema *domainRLS.AuthSchema,
) error {
	variablesJSON, err := json.Marshal(authSchema.Variables)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	err = r.q.UpsertProjectAuthSchema(ctx, dbgen.UpsertProjectAuthSchemaParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Variables:   variablesJSON,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// DeleteByProjectID deletes the auth schema for a given project.
func (r *SqlAuthSchemaRepository) DeleteByProjectID(ctx context.Context, orgName, projectSlug string) error {
	err := r.q.DeleteProjectAuthSchema(ctx, dbgen.DeleteProjectAuthSchemaParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// Compile-time interface check.
var _ AuthSchemaRepository = (*SqlAuthSchemaRepository)(nil)

// AuthSchemaToDomain converts a dbgen.ProjectAuthSchema to a domain AuthSchema.
// This helper function is provided for use by other layers.
func AuthSchemaToDomain(row dbgen.ProjectAuthSchema) (*domainRLS.AuthSchema, error) {
	var variables []domainRLS.AuthVariable
	if err := json.Unmarshal(row.Variables, &variables); err != nil {
		return nil, err
	}

	return &domainRLS.AuthSchema{
		ProjectID: row.ProjectSlug,
		Variables: variables,
	}, nil
}

// AuthSchemaToCreateParams converts a domain AuthSchema to dbgen upsert params.
// This helper function is provided for use by other layers.
func AuthSchemaToCreateParams(
	orgName, projectSlug string,
	authSchema *domainRLS.AuthSchema,
) (dbgen.UpsertProjectAuthSchemaParams, error) {
	variablesJSON, err := json.Marshal(authSchema.Variables)
	if err != nil {
		return dbgen.UpsertProjectAuthSchemaParams{}, err
	}

	return dbgen.UpsertProjectAuthSchemaParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Variables:   variablesJSON,
	}, nil
}
