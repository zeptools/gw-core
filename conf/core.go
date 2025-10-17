package conf

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/zeptools/gw-core/db"
	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/db/sqldb"
)

type Core struct {
	AppName            string                     `json:"app_name"`
	AppRoot            string                     `json:"-"` // filled from compiled paths
	Listen             string                     `json:"listen"`
	Host               string                     `json:"host"` // can be used to generate public url endpoints
	Context            context.Context            `json:"-"`
	VolatileKV         *sync.Map                  `json:"-"`
	DBConf             CommonDBConf               `json:"-"` // Init manually. e.g. for separate file
	KVDBClient         kvdb.Client                `json:"-"`
	MainDBClient       sqldb.Client               `json:"-"`
	MainDBRawStore     *sqldb.RawStore            `json:"-"`
	MainDBPlaceholder  func(...int) string        `json:"-"`
	MainDBPlaceholders func(int, ...int) []string `json:"-"`
	HttpClient         *http.Client               `json:"-"`
	SessionLocks       *sync.Map                  `json:"-"`          // map[string]*sync.Mutex
	DebugOpts          DebugOpts                  `json:"debug_opts"` // Do not promote
}

func (c *Core) CleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")

	// clean up DB clients
	db.CloseClient("KVDBClient", c.KVDBClient)
	db.CloseClient("MainDBClient", c.MainDBClient)

	log.Println("[INFO] App Resource Cleanup Complete")
}

type CommonDBConf struct {
	KV   kvdb.Conf  `json:"kv"`
	Main sqldb.Conf `json:"main"`
}

type DebugOpts struct {
	MaintenanceMode int `json:"maintenance_mode"`
	AuthBreak       int `json:"auth_break"`
}
