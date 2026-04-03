package repository

import (
	"database/sql"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/config"
	"time"

	// Register MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// NewSQLConnection creates a *sql.DB connection using the standard library MySQL driver.
func NewSQLConnection(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, bizerrors.Errorf("failed to open MySQL connection: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, bizerrors.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
