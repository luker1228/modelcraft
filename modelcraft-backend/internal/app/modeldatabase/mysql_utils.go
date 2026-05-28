package modeldatabase

import (
	"context"
	"database/sql"
)

var systemDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
}

// listMySQLDatabases returns all non-system database names visible in the given connection.
func listMySQLDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close() //nolint:errcheck
			return nil, err
		}
		if !systemDatabases[name] {
			names = append(names, name)
		}
	}
	rows.Close() //nolint:errcheck
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
