package sqldb

import (
	"context"
)

type Client interface {
	Init() error
	Open(ctx context.Context) error
	Close() error
	GetHandle() Handle
	Handle // Methods required for Handle are also required, so, promote it
	GetConf() *Conf
	GetDSN() string
	Ping(ctx context.Context) error
	BeginTx(ctx context.Context) (Tx, error)
}
