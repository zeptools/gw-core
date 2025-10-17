package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeptools/gw-core/db/sqldb"
)

type DBHandle struct {
	// sqldb.DBHandle // [Interface]

	db *sql.DB
}

// Ensure mysql.DBHandle implements sqldb.DBHandle interface
var _ sqldb.DBHandle = (*DBHandle)(nil)

func (h *DBHandle) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	result, err := h.db.ExecContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Result{result: result}, nil
}

func (h *DBHandle) QueryRows(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	rows, err := h.db.QueryContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

func (h *DBHandle) QueryRow(ctx context.Context, query string, args ...any) sqldb.Row {
	row := h.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}
}

// CopyFrom - params: table, columns, rows
func (h *DBHandle) CopyFrom(_ context.Context, _ string, _ []string, _ [][]any) (int64, error) {
	// MySQL doesn't have native COPY
	// ToDo: emulate batch insert
	return 0, fmt.Errorf("method `CopyFrom` not supported for MySQL")
}

// Listen - param: channel
func (h *DBHandle) Listen(_ context.Context, _ string) (<-chan sqldb.Notification, error) {
	return nil, fmt.Errorf("method `Listen` not supported for MySQL")
}

func (h *DBHandle) InsertStmt(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	trimmed := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "INSERT") {
		return nil, fmt.Errorf("InsertStmt must start with INSERT")
	}
	result, err := h.db.ExecContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Result{result: result}, nil
}

func (h *DBHandle) Prepare(ctx context.Context, query string) (sqldb.PreparedStmt, error) {
	stmt, err := h.db.PrepareContext(ctx, query)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &PreparedStmt{stmt: stmt}, nil
}
