package sqldb

import (
	"context"
)

type Client interface {
	DBHandle() Handle
	Conf() *Conf
	DSN() string
	RawSQLStore() *RawStore

	Init() error
	Open(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error

	Handle // Handle Methods are also required, so, promote it

	BeginTx(ctx context.Context) (Tx, error)
}
