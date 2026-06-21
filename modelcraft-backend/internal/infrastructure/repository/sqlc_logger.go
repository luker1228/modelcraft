package repository

import (
	"context"
	"database/sql"
	"modelcraft/pkg/logfacade"
	"strings"
	"time"
)

// SqlcLogLevel controls the verbosity of sqlc query logging.
// Mirrors GORM's log level conventions: Silent, Error, Warn, Info.
type SqlcLogLevel int

const (
	// SqlcLogSilent disables all sqlc query logging.
	SqlcLogSilent SqlcLogLevel = iota
	// SqlcLogError logs only queries that return an error.
	SqlcLogError
	// SqlcLogWarn logs errors and queries that exceed the slow query threshold.
	SqlcLogWarn
	// SqlcLogInfo logs all queries regardless of outcome.
	SqlcLogInfo
)

// sqlcLogger wraps a DBTX and logs SQL execution using logfacade.
// It implements the dbgen.DBTX interface so it can be passed directly to dbgen.New().
//
// QueryRowContext cannot report elapsed time or errors at call time because
// *sql.Row defers execution until Scan() is called. Only the query dispatch
// is logged for that method.
type sqlcLogger struct {
	db            DBTX
	level         SqlcLogLevel
	slowThreshold time.Duration
}

// DBTX mirrors the dbgen.DBTX interface locally to avoid an import cycle.
// It is satisfied by *sql.DB, *sql.Tx, and sqlcLogger itself.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// NewSqlcLogger wraps db with a logging layer.
// level controls which events are emitted; slowThreshold triggers Warn-level
// logs when a query exceeds the duration (0 disables slow-query detection).
func NewSqlcLogger(db DBTX, level SqlcLogLevel, slowThreshold time.Duration) DBTX {
	return &sqlcLogger{
		db:            db,
		level:         level,
		slowThreshold: slowThreshold,
	}
}

// ExecContext executes a query and logs the outcome with elapsed time.
func (l *sqlcLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if l.level == SqlcLogSilent {
		return l.db.ExecContext(ctx, query, args...)
	}

	start := time.Now()
	result, err := l.db.ExecContext(ctx, query, args...)
	elapsed := time.Since(start)

	l.log(ctx, query, args, -1, elapsed, err)
	return result, err
}

// PrepareContext prepares a statement. No SQL logging is performed here
// because no data is transferred yet.
func (l *sqlcLogger) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return l.db.PrepareContext(ctx, query)
}

// QueryContext executes a query returning multiple rows and logs the outcome.
func (l *sqlcLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if l.level == SqlcLogSilent {
		return l.db.QueryContext(ctx, query, args...)
	}

	start := time.Now()
	rows, err := l.db.QueryContext(ctx, query, args...)
	elapsed := time.Since(start)

	l.log(ctx, query, args, -1, elapsed, err)
	return rows, err
}

// QueryRowContext dispatches a single-row query.
// Because *sql.Row defers execution until Scan(), only the dispatch is logged
// without elapsed time or row count.
func (l *sqlcLogger) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if l.level == SqlcLogInfo {
		logfacade.GetLogger(ctx).With(
			logfacade.String(logfacade.SQLKey, cleanSQL(query)),
			logfacade.Any(logfacade.SQLArgsKey, args),
		).Infof(ctx, "[SQLC] query row dispatched")
	}
	return l.db.QueryRowContext(ctx, query, args...)
}

// log emits a log entry based on the configured level and query outcome.
// rows is -1 when the caller cannot determine the affected row count.
func (l *sqlcLogger) log(
	ctx context.Context, query string, args []interface{}, rows int64, elapsed time.Duration, err error,
) {
	logger := logfacade.GetLogger(ctx)
	cleanedSQL := cleanSQL(query)

	switch {
	case err != nil && l.level >= SqlcLogError:
		logger.With(
			logfacade.Err(err),
			logfacade.String(logfacade.SQLKey, cleanedSQL),
			logfacade.Any(logfacade.SQLArgsKey, args),
			logfacade.Duration(logfacade.ElapsedKey, elapsed),
		).Errorf(ctx, nil, "[SQLC] query error")

	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= SqlcLogWarn:
		logger.With(
			logfacade.String(logfacade.SQLKey, cleanedSQL),
			logfacade.Any(logfacade.SQLArgsKey, args),
			logfacade.Duration(logfacade.ElapsedKey, elapsed),
			logfacade.Duration(logfacade.ThresholdKey, l.slowThreshold),
		).Warnf(ctx, "[SQLC] slow query")

	case l.level >= SqlcLogInfo:
		fields := []logfacade.Field{
			logfacade.String(logfacade.SQLKey, cleanedSQL),
			logfacade.Any(logfacade.SQLArgsKey, args),
			logfacade.Duration(logfacade.ElapsedKey, elapsed),
		}
		if rows >= 0 {
			fields = append(fields, logfacade.Int64(logfacade.RowsKey, rows))
		}
		logger.With(fields...).Infof(ctx, "[SQLC] query ok")
	}
}

// cleanSQL collapses newlines and redundant whitespace in a SQL string into a
// single space so that multi-line sqlc query constants are logged on one line.
func cleanSQL(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
