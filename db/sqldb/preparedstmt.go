package sqldb

import "context"

type PreparedStmt interface {
	Query(ctx context.Context, args ...any) (Rows, error)
	Exec(ctx context.Context, args ...any) (Result, error)
	Close() error
}
