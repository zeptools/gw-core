package mysql

import (
	"database/sql"
	"errors"

	"github.com/zeptools/gw-core/db/sqldb"
)

type Row struct {
	row *sql.Row
}

// Ensure mysql.Row implements sqldb.Row interface
var _ sqldb.Row = (*Row)(nil)

func (r *Row) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if errors.Is(err, sql.ErrNoRows) {
		return sqldb.ErrNoRows
	}
	return err
}
