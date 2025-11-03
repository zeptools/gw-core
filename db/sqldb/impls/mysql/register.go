package mysql

import (
	"github.com/zeptools/gw-core/db/sqldb"
)

func Register() {
	sqldb.RegisterFactory(DBType, NewClient)
}
