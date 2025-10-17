package pgsql

import (
	"context"

	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreparedStmt struct {
	conn     *pgxpool.Conn
	stmtName string
}

// Ensure pgsql.PreparedStmt implements sqldb.PreparedStmt interface
var _ sqldb.PreparedStmt = (*PreparedStmt)(nil)

func (p *PreparedStmt) Query(ctx context.Context, args ...any) (sqldb.Rows, error) {
	rows, err := p.conn.Query(ctx, p.stmtName, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{current: rows}, nil
}

func (p *PreparedStmt) Exec(ctx context.Context, args ...any) (sqldb.Result, error) {
	tag, err := p.conn.Exec(ctx, p.stmtName, args...)
	if err != nil {
		return nil, err
	}
	return &Result{tag: tag}, nil
}

func (p *PreparedStmt) Close() error {
	p.conn.Release()
	return nil
}
