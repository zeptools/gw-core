package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeptools/gw-core/db/sqldb"
)

type Handle struct {
	*sql.DB // [Embedded]
}

// Ensure mysql.Handle implements sqldb.Handle interface
var _ sqldb.Handle = (*Handle)(nil)

func (h *Handle) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	result, err := h.DB.ExecContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Result{result: result}, nil
}

func (h *Handle) QueryRows(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	rows, err := h.DB.QueryContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

func (h *Handle) QueryRow(ctx context.Context, query string, args ...any) sqldb.Row {
	row := h.DB.QueryRowContext(ctx, query, args...)
	return &Row{row: row}
}

// CopyFrom - params: table, columns, rows
func (h *Handle) CopyFrom(_ context.Context, _ string, _ []string, _ [][]any) (int64, error) {
	// MySQL doesn't have native COPY
	// ToDo: emulate batch insert
	return 0, fmt.Errorf("method `CopyFrom` not supported for MySQL")
}

// Listen - param: channel
func (h *Handle) Listen(_ context.Context, _ string) (<-chan sqldb.Notification, error) {
	return nil, fmt.Errorf("method `Listen` not supported for MySQL")
}

func (h *Handle) InsertStmt(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	trimmed := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "INSERT") {
		return nil, fmt.Errorf("InsertStmt must start with INSERT")
	}
	result, err := h.DB.ExecContext(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Result{result: result}, nil
}

func (h *Handle) Prepare(ctx context.Context, query string) (sqldb.PreparedStmt, error) {
	stmt, err := h.DB.PrepareContext(ctx, query)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &PreparedStmt{stmt: stmt}, nil
}
