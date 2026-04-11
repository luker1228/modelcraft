package dbgenwrap

import "modelcraft/internal/infrastructure/sqlerr"

// WrapSQLErrorInPlace 将 SQL 错误包装为 RepositoryError。
func WrapSQLErrorInPlace(err *error) {
	sqlerr.WrapSQLErrorInPlace(err)
}
