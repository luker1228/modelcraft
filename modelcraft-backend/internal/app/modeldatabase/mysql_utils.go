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
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if !systemDatabases[name] {
			names = append(names, name)
		}
	}
	return names, rows.Err()
}
