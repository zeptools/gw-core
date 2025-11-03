package pgsql

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeptools/gw-core/db/sqldb"
)

type Handle struct {
	*pgxpool.Pool // [Embedded]
}

var _ sqldb.Handle = (*Handle)(nil)

func (h *Handle) Exec(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	tag, err := h.Pool.Exec(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Result{tag: tag}, nil
}

func (h *Handle) QueryRows(ctx context.Context, query string, args ...any) (sqldb.Rows, error) {
	rows, err := h.Pool.Query(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	return &Rows{
		conn:    nil, // Pool manages connection, no need to release here
		current: rows,
		batch:   nil, // single query, no batch
	}, nil
}

func (h *Handle) QueryRow(ctx context.Context, query string, args ...any) sqldb.Row {
	row := h.Pool.QueryRow(ctx, query, args...)
	return &Row{row: row}
}

func (h *Handle) CopyFrom(ctx context.Context, table string, columns []string, rows [][]any) (int64, error) {
	src := pgx.CopyFromRows(rows)
	count, err := h.Pool.CopyFrom(ctx, pgx.Identifier{table}, columns, src)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	return count, err
}

func (h *Handle) Listen(ctx context.Context, channel string) (<-chan sqldb.Notification, error) {
	conn, err := h.Pool.Acquire(ctx)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}

	notifyCh := make(chan sqldb.Notification)

	go func() {
		defer conn.Release()
		defer close(notifyCh)

		_, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", pgx.Identifier{channel}.Sanitize()))
		if err != nil {
			log.Printf("[WARN] failed to LISTEN on %s: %v", channel, err)
			return
		}

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				log.Printf("[WARN] Listen loop ended for %s: %v", channel, err)
				return
			}
			select {
			case notifyCh <- sqldb.Notification{
				Channel: notification.Channel,
				Payload: notification.Payload,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return notifyCh, nil
}

func (h *Handle) InsertStmt(ctx context.Context, query string, args ...any) (sqldb.Result, error) {
	trimmed := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "INSERT") {
		return nil, fmt.Errorf("InsertStmt must start with INSERT")
	}
	// append RETURNING id if missing
	if !strings.Contains(strings.ToUpper(query), "RETURNING") {
		query += " RETURNING id"
		var id int64
		err := h.Pool.QueryRow(ctx, query, args...).Scan(&id)
		if err != nil {
			return nil, err
		}
		return &Result{lastInsertID: id}, nil
	}

	tag, err := h.Pool.Exec(ctx, query, args...)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	return &Result{tag: tag}, err
}

func (h *Handle) Prepare(ctx context.Context, query string) (sqldb.PreparedStmt, error) {
	conn, err := h.Pool.Acquire(ctx)
	// NOTE: We can process a DBMS-specific error to produce a better abstracted error
	if err != nil {
		return nil, err
	}
	stmtName := fmt.Sprintf("stmt_%x", time.Now().UnixNano())
	_, err = conn.Conn().Prepare(ctx, stmtName, query)
	if err != nil {
		conn.Release()
		return nil, err
	}
	return &PreparedStmt{conn: conn, stmtName: stmtName}, nil
}
