package rls

import (
	"context"
	"database/sql"
	domainrls "modelcraft/internal/domain/rls"
	"modelcraft/internal/infrastructure/dbgen"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSqlPolicyRepository_Upsert_PersistsPlainTextExpressions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSqlPolicyRepository(dbgen.New(db))
	policy := &domainrls.Policy{
		ModelID:       "model-1",
		PolicyName:    "owner_only",
		Action:        domainrls.ActionRead,
		Role:          "admin",
		UsingExpr:     `row.owner_id == auth.userid`,
		WithCheckExpr: `input.owner_id == auth.userid`,
	}

	mock.ExpectExec("INSERT INTO model_rls_policies").
		WithArgs(
			"org-a",
			"project-a",
			"model-1",
			"owner_only",
			dbgen.ModelRlsPoliciesAction(domainrls.ActionRead),
			"admin",
			sql.NullString{String: "row.owner_id == auth.userid", Valid: true},
			sql.NullString{String: "input.owner_id == auth.userid", Valid: true},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Upsert(context.Background(), "org-a", "project-a", policy)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
