package pgsql

import (
	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Rows struct {
	conn    *pgxpool.Conn
	current pgx.Rows
	batch   pgx.BatchResults
}

// Ensure pgsql.Rows implements sqldb.Rows
var _ sqldb.Rows = (*Rows)(nil)

func (r *Rows) Next() bool {
	return r.current.Next()
}

func (r *Rows) Scan(dest ...any) error {
	raw := make([]any, len(dest))
	for i, d := range dest {
		switch d.(type) {
		case *bool:
			raw[i] = new(int16)
		default:
			raw[i] = d
		}
	}
	if err := r.current.Scan(raw...); err != nil {
		return err
	}
	for i, d := range dest {
		switch v := d.(type) {
		case *bool:
			*v = *(raw[i].(*int16)) != 0
		}
	}
	return nil
}

func (r *Rows) Close() error {
	if r.current != nil {
		r.current.Close()
	}
	if r.batch != nil {
		_ = r.batch.Close()
	}
	if r.conn != nil {
		r.conn.Release()
	}
	return nil
}

func (r *Rows) Err() error {
	return r.current.Err()
}

func (r *Rows) NextResultSet() bool {
	if r.batch == nil {
		return false
	}
	nextRows, err := r.batch.Query()
	if err != nil {
		// No more result sets or query failed
		return false
	}
	r.current = nextRows
	return true
}
