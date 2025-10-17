package mysql

import (
	"database/sql"

	"github.com/zeptools/gw-core/db/sqldb"
)

type Rows struct {
	rows *sql.Rows
}

// Ensure mysql.Rows implements sqldb.Rows interface
var _ sqldb.Rows = (*Rows)(nil)

func (r *Rows) Next() bool {
	return r.rows.Next()
}

func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *Rows) Close() error {
	return r.rows.Close()
}

func (r *Rows) NextResultSet() bool {
	// MySQL via database/sql doesn't support multiple result sets here
	return false
}

func (r *Rows) Err() error {
	return r.rows.Err()
}
