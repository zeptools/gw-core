package pgsql

import (
	"fmt"

	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/jackc/pgx/v5/pgconn"
)

type Result struct {
	tag          pgconn.CommandTag
	lastInsertID int64 // Auto-increment ID
}

// Ensure pgsql.Result implements sqldb.Result
var _ sqldb.Result = (*Result)(nil)

func (r *Result) RowsAffected() (int64, error) {
	return r.tag.RowsAffected(), nil
}

// LastInsertId - PostgreSQL does not support LastInsertId.
func (r *Result) LastInsertId() (int64, error) {
	if r.lastInsertID != 0 {
		return r.lastInsertID, nil
	}
	// err := dbHandle.QueryRow(ctx, "INSERT INTO users(first_name, last_name) VALUES($1, $2) RETURNING id", "John", "Doe").Scan(&id)
	return 0, fmt.Errorf("LastInsertId not supported; use `RETURNING id` instead")
}
