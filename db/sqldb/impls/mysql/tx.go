package mysql

import (
	"context"
	"database/sql"

	"github.com/zeptools/gw-core/db/sqldb"
)

type Tx struct {
	tx *sql.Tx
}

// Ensure mysql.Tx implements sqldb.Tx interface
var _ sqldb.Tx = (*Tx)(nil)

func (t *Tx) Commit(_ context.Context) error {
	return t.tx.Commit()
}

func (t *Tx) Rollback(_ context.Context) error {
	return t.tx.Rollback()
}

func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Query(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}
