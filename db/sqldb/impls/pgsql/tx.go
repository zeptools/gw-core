package pgsql

import (
	"context"

	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/jackc/pgx/v5"
)

type Tx struct {
	tx pgx.Tx
}

// Ensure pgsql.Tx implements sqldb.Tx
var _ sqldb.Tx = (*Tx)(nil)

func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	tag, err := t.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Result{tag: tag}, nil
}

func (t *Tx) Query(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{
		conn:    nil, // tx already owns the connection
		current: rows,
		batch:   nil,
	}, nil
}
