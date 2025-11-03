package sqldb

import (
	"context"
)

type Client interface {
	GetHandle() Handle
	GetConf() *Conf
	GetDSN() string
	GetRawSQLStore() *RawStore

	Init() error
	Open(ctx context.Context) error
	Ping(ctx context.Context) error
	Close() error

	Handle // Handle Methods are also required, so, promote it

	BeginTx(ctx context.Context) (Tx, error)
}
