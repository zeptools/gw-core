package conf

import (
	"context"
	"encoding/json/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/zeptools/gw-core/db"
	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/zeptools/gw-core/schedjobs"
	"github.com/zeptools/gw-core/storages"
	"github.com/zeptools/gw-core/svc"
	"github.com/zeptools/gw-core/throttle"
	"github.com/zeptools/gw-core/uds"
	"github.com/zeptools/gw-core/web"
)

// Core - common config
// SU = Type for Session User _ e.g. string, int64, etc
type Core[SU comparable] struct {
	AppName             string                     `json:"app_name"`
	Listen              string                     `json:"listen"`     // HTTP Server Listen IP:PORT Address
	Host                string                     `json:"host"`       // HTTP Host. Can be used to generate public url endpoints
	DebugOpts           DebugOpts                  `json:"debug_opts"` // Debug Options
	AppRoot             string                     `json:"-"`          // Filled from compiled paths
	RootCtx             context.Context            `json:"-"`          // Global Context with RootCancel
	RootCancel          context.CancelFunc         `json:"-"`          // CancelFunc for RootCtx
	UDSService          *uds.Service               `json:"-"`          // PrepareUDSService()
	JobScheduler        *schedjobs.Scheduler       `json:"-"`          // PrepareJobScheduler()
	WebService          *web.Service               `json:"-"`
	ThrottleBucketStore *throttle.BucketStore[SU]  `json:"-"` // PrepareThrottleBucketStore()
	VolatileKV          *sync.Map                  `json:"-"` // map[string]string
	SessionLocks        *sync.Map                  `json:"-"` // map[string]*sync.Mutex
	ActionLocks         *sync.Map                  `json:"-"` // map[string]struct{}
	StorageConf         storages.Conf              `json:"-"` // LoadStorageConf()
	DBConf              CommonDBConf               `json:"-"` // LoadDBConf()
	HttpClient          *http.Client               `json:"-"` // for requests to external apis
	KVDBClient          kvdb.Client                `json:"-"`
	MainDBClient        sqldb.Client               `json:"-"`
	MainDBRawStore      *sqldb.RawStore            `json:"-"`
	MainDBPlaceholder   func(...int) string        `json:"-"`
	MainDBPlaceholders  func(int, ...int) []string `json:"-"`
	services            []svc.Service              // Services to Manage
	done                chan error
}

// BaseInit - 1st step for initialization
// 1. set AppRoot
// 2. load config/.core.json file
// 3. prepare base fields
// 4. Start ShutdownSignalListener
func (c *Core[SU]) BaseInit(appRoot string, rootCtx context.Context, rootCancel context.CancelFunc) error {
	c.AppRoot = appRoot
	// Load .env.json
	envFilePath := filepath.Join(appRoot, "config", ".core.json")
	//file, readErr := os.Open(envFilePath) // (*os.File, error)
	envBytes, err := os.ReadFile(envFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(envBytes, c); err != nil {
		return err
	}
	c.RootCtx = rootCtx
	c.RootCancel = rootCancel
	c.prepareDefaultFeatures()
	c.startShutdownSignalListener()
	return nil
}

func (c *Core[SU]) prepareDefaultFeatures() {
	c.VolatileKV = &sync.Map{}
	c.SessionLocks = &sync.Map{}
	c.HttpClient = &http.Client{}
	c.ActionLocks = &sync.Map{}
}

func (c *Core[SU]) AddService(s svc.Service) {
	c.services = append(c.services, s)
}

func (c *Core[SU]) StartServices() error {
	c.done = make(chan error, len(c.services))
	for _, s := range c.services {
		err := s.Start()
		if err != nil {
			return err
		}
		go func(s svc.Service) {
			err := <-s.Done()
			c.done <- err
		}(s) // pass the loop var to the param. otherwise, they are captured inside goroutine lazily
	}
	return nil
}

func (c *Core[SU]) WaitServicesDone() error {
	for i := 0; i < len(c.services); i++ {
		if err := <-c.done; err != nil {
			return err
		}
	}
	return nil
}

func (c *Core[SU]) StopServices() {
	for _, s := range c.services {
		s.Stop()
	}
}

var once sync.Once

func (c *Core[SU]) startShutdownSignalListener() {
	once.Do(func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigs
			log.Printf("[INFO] got signal [%s]. shutting down app [%s] ...", sig, c.AppName)
			c.RootCancel() // broadcast to all child services via Context.Done()
		}()
	})
	log.Printf("[INFO][CORE] shutdown signal listener started")
}

func (c *Core[SU]) PrepareWebService(addr string, router http.Handler) {
	c.WebService = web.NewService(c.RootCtx, addr, router)
	c.AddService(c.WebService)
}

func (c *Core[SU]) PrepareUDSService(sockPath string, cmdMap map[string]uds.CmdHnd) {
	c.UDSService = uds.NewService(c.RootCtx, sockPath, cmdMap)
	c.AddService(c.UDSService)
}

func (c *Core[SU]) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler(c.RootCtx)
	c.AddService(c.JobScheduler)
}

func (c *Core[SU]) PrepareThrottleBucketStore(cleanupCycle time.Duration, cleanupOlderThan time.Duration) {
	c.ThrottleBucketStore = throttle.NewBucketStore[SU](c.RootCtx, cleanupCycle, cleanupOlderThan)
	c.AddService(c.ThrottleBucketStore)
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

func (c *Core[SU]) ResourceCleanUp() {
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
