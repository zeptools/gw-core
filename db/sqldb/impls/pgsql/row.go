package pgsql

import (
	"errors"

	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/jackc/pgx/v5"
)

type Row struct {
	row pgx.Row
}

// Ensure pgsql.Row implements sqldb.Row interface
var _ sqldb.Row = (*Row)(nil)

func (r *Row) Scan(dest ...any) error {
	// first, scan to `int16`s instead of `bool`s
	raw := make([]any, len(dest))
	for i, d := range dest {
		switch d.(type) {
		case *bool:
			// scan into a temporary int16
			raw[i] = new(int16)
		default:
			raw[i] = d
		}
	}
	err := r.row.Scan(raw...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqldb.ErrNoRows
		}
		return err
	}
	// fill dest with `bool` as `bool`
	for i, d := range dest {
		switch v := d.(type) {
		case *bool:
			*v = *(raw[i].(*int16)) != 0
		}
	}
	return nil
}
