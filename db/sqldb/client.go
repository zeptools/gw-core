package sqldb

import (
	"context"

	"github.com/zeptools/gw-core/db"
)

type Client interface {
	db.Client[DBHandle]

	BeginTx(ctx context.Context) (Tx, error)
}
