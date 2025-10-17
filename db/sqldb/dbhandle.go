package sqldb

import "context"

type DBHandle interface {
	// Exec executes SQL statement like INSERT, UPDATE, DELETE.
	Exec(ctx context.Context, query string, args ...any) (Result, error) // Executes General SQL Statement(s)

	QueryRows(ctx context.Context, query string, args ...any) (Rows, error) // Eager. Fail upfront on statement execution
	QueryRow(ctx context.Context, query string, args ...any) Row            // Lazy. only fails at Scan()

	// CopyFrom inserts many rows at once into a table, not executing individual INSERT statements.
	// You stream all the rows in one operation
	CopyFrom(ctx context.Context, table string, columns []string, rows [][]any) (int64, error)

	Listen(ctx context.Context, channel string) (<-chan Notification, error)
	Prepare(ctx context.Context, query string) (PreparedStmt, error)

	// InsertStmt - Single INSERT statement, placeholders only
	// to guarantee Result.LastInsertedId() works for auto-increment `id`
	InsertStmt(ctx context.Context, query string, args ...any) (Result, error)
}
