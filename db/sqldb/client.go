package sqldb

import (
	"context"
)

type Client interface {
	DBHandle() Handle
	Conf() *Conf
	DSN() string

	SinglePlaceholder(nth ...int) string       // n'th Placeholder (Optional, Default = 1)
	Placeholders(cnt int, start ...int) string // Count, start (Optional, Default = 1)
	RawSQLStore() *RawSQLStore

	Init() error
	Open(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error

	Handle // Handle Methods are also required, so, promote it

	BeginTx(ctx context.Context) (Tx, error)
}
