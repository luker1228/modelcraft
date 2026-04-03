package repository

import (
	"database/sql"
)

// ConnectionFactory holds database connections for repository construction.
type ConnectionFactory struct {
	// SqlDB is the *sql.DB connection used by sqlc-based repositories.
	SqlDB *sql.DB
}

// NewConnectionFactory creates a ConnectionFactory from a *sql.DB connection.
func NewConnectionFactory(sqlDB *sql.DB) *ConnectionFactory {
	return &ConnectionFactory{
		SqlDB: sqlDB,
	}
}
