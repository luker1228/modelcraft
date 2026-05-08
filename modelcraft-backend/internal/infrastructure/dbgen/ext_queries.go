package dbgen

// DB returns the underlying DBTX held by the Queries instance.
// Use this when you need to pass the raw connection (e.g. to a repository
// that operates outside the sqlc-generated query set) within a transaction.
func (q *Queries) DB() DBTX {
	return q.db
}
