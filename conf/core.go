package conf

import (
	"context"
	"encoding/json/v2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeptools/gw-core/db"
	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/zeptools/gw-core/schedjobs"
	"github.com/zeptools/gw-core/storages"
	"github.com/zeptools/gw-core/throttle"
)

// Core - common config
// SU = Type for Session User _ e.g. string, int64, etc
type Core[SU comparable] struct {
	AppName             string                     `json:"app_name"`
	Listen              string                     `json:"listen"`
	AppRoot             string                     `json:"-"`    // Filled from compiled paths
	Host                string                     `json:"host"` // Can be used to generate public url endpoints
	Context             context.Context            `json:"-"`    // Shared Context
	JobScheduler        *schedjobs.Scheduler       `json:"-"`    // PrepareJobScheduler()
	ThrottleBucketStore *throttle.BucketStore[SU]  `json:"-"`    // PrepareThrottleBucketStore()
	VolatileKV          *sync.Map                  `json:"-"`    // map[string]string
	SessionLocks        *sync.Map                  `json:"-"`    // map[string]*sync.Mutex
	ActionLocks         *sync.Map                  `json:"-"`    // map[string]struct{}
	StorageConf         storages.Conf              `json:"-"`    // LoadStorageConf()
	DBConf              CommonDBConf               `json:"-"`    // LoadDBConf()
	KVDBClient          kvdb.Client                `json:"-"`
	MainDBClient        sqldb.Client               `json:"-"`
	MainDBRawStore      *sqldb.RawStore            `json:"-"`
	MainDBPlaceholder   func(...int) string        `json:"-"`
	MainDBPlaceholders  func(int, ...int) []string `json:"-"`
	HttpClient          *http.Client               `json:"-"`
	DebugOpts           DebugOpts                  `json:"debug_opts"`
}

func (c *Core[SU]) InitLoadEnvFile(appRoot string, ctx context.Context) error {
	c.AppRoot = appRoot
	c.Context = ctx
	// Load .env.json
	envFilePath := filepath.Join(c.AppRoot, "config", ".env.json")
	//file, readErr := os.Open(envFilePath) // (*os.File, error)
	envBytes, err := os.ReadFile(envFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(envBytes, c); err != nil {
		return err
	}
	return nil
}

func (c *Core[SU]) PrepareBase() {
	c.VolatileKV = &sync.Map{}
	c.SessionLocks = &sync.Map{}
	c.HttpClient = &http.Client{}
	c.ActionLocks = &sync.Map{}
}

func (c *Core[SU]) PrepareThrottleBucketStore() {
	c.ThrottleBucketStore = throttle.NewBucketStore[SU]()
}

func (c *Core[SU]) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler()
}

func (c *Core[SU]) PrepareMainDBRawStore() {
	c.MainDBRawStore = sqldb.NewRawStore()
}

func (c *Core[SU]) LoadDBConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".databases.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.DBConf); err != nil {
		return err
	}
	return nil
}

func (c *Core[SU]) LoadStorageConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".storages.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.StorageConf); err != nil {
		return err
	}
	return nil
}

func (c *Core[SU]) CleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")
	// Clean up DB clients ----
	db.CloseClient("KVDBClient", c.KVDBClient)
	db.CloseClient("MainDBClient", c.MainDBClient)
	//----
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
