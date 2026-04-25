package repository

import (
	"context"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"testing"

	sqldriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSqlEndUserRepository_Save_DuplicateKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSqlEndUserRepository(db, "org-a", "project-a")

	user := &enduser.EndUser{
		ID:          "user-1",
		OrgName:     "org-a",
		Username:    "alice",
		Password:    enduser.NewHashedPasswordFromHash("hashed-password"),
		IsForbidden: false,
		CreatedBy:   "dev-1",
	}

	mock.ExpectExec("INSERT INTO end_user_users").
		WithArgs(
			user.ID,
			user.OrgName,
			user.Username,
			user.Password.Hash,
			0,
			user.CreatedBy,
		).
		WillReturnError(&sqldriver.MySQLError{Number: 1062, Message: "Duplicate entry"})

	err = repo.Save(context.Background(), user)
	require.Error(t, err)
	assert.True(t, shared.IsDuplicateKeyError(err))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlEndUserRepository_GetByUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSqlEndUserRepository(db, "org-a", "project-a")

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"password",
		"is_forbidden",
		"created_by",
		"created_at",
		"updated_at",
	})

	mock.ExpectQuery("SELECT id, username, password, is_forbidden, created_by, created_at, updated_at").
		WithArgs("alice", "org-a").
		WillReturnRows(rows)

	user, err := repo.GetByUsername(context.Background(), "org-a", "alice")
	require.NoError(t, err)
	assert.Nil(t, user)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlEndUserRepository_Delete_NoRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSqlEndUserRepository(db, "org-a", "project-a")

	mock.ExpectExec("DELETE FROM end_user_users").
		WithArgs("user-404", "org-a").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), "org-a", "user-404")
	require.Error(t, err)
	assert.True(t, shared.IsRepoError(err, shared.ErrTypeNoRowsAffected))
	require.NoError(t, mock.ExpectationsWereMet())
}
