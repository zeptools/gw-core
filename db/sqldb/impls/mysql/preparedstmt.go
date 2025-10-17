package mysql

import (
	"context"
	"database/sql"

	"github.com/zeptools/gw-core/db/sqldb"
)

type PreparedStmt struct {
	stmt *sql.Stmt
}

// Ensure mysql.PreparedStmt implements sqldb.PreparedStmt interface
var _ sqldb.PreparedStmt = (*PreparedStmt)(nil)

func (p *PreparedStmt) Query(ctx context.Context, args ...any) (sqldb.Rows, error) {
	return p.stmt.QueryContext(ctx, args...)
}

func (p *PreparedStmt) Exec(ctx context.Context, args ...any) (sqldb.Result, error) {
	return p.stmt.ExecContext(ctx, args...)
}

func (p *PreparedStmt) Close() error {
	return p.stmt.Close()
}
