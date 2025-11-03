package conf

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/db/kvdb/impls/redis"
	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/zeptools/gw-core/db/sqldb/impls/mysql"
	"github.com/zeptools/gw-core/db/sqldb/impls/pgsql"
	"github.com/zeptools/gw-core/schedjobs"
	"github.com/zeptools/gw-core/storages"
	"github.com/zeptools/gw-core/svc"
	"github.com/zeptools/gw-core/throttle"
	"github.com/zeptools/gw-core/uds"
	"github.com/zeptools/gw-core/web"
)

// Core - common config
// B = BucketID Type _ e.g. string, int64, etc
type Core[B comparable] struct {
	AppName             string                   `json:"app_name"`
	Listen              string                   `json:"listen"`     // HTTP Server Listen IP:PORT Address
	Host                string                   `json:"host"`       // HTTP Host. Can be used to generate public url endpoints
	DebugOpts           DebugOpts                `json:"debug_opts"` // Debug Options
	AppRoot             string                   `json:"-"`          // Filled from compiled paths
	RootCtx             context.Context          `json:"-"`          // Global Context with RootCancel
	RootCancel          context.CancelFunc       `json:"-"`          // CancelFunc for RootCtx
	UDSService          *uds.Service             `json:"-"`          // PrepareUDSService()
	JobScheduler        *schedjobs.Scheduler     `json:"-"`          // PrepareJobScheduler()
	WebService          *web.Service             `json:"-"`
	ThrottleBucketStore *throttle.BucketStore[B] `json:"-"` // PrepareThrottleBucketStore()
	VolatileKV          *sync.Map                `json:"-"` // map[string]string
	SessionLocks        *sync.Map                `json:"-"` // map[string]*sync.Mutex
	ActionLocks         *sync.Map                `json:"-"` // map[string]struct{}
	StorageConf         storages.Conf            `json:"-"` // LoadStorageConf()
	HttpClient          *http.Client             `json:"-"` // for requests to external apis
	KVDBConf            kvdb.Conf                `json:"-"` // LoadKVDBConf()
	KVDBClient          kvdb.Client              `json:"-"` // PrepareKVDBClient()
	SQLDBConfs          map[string]*sqldb.Conf   `json:"-"` // LoadSQLDBConfs()
	SQLDBClients        map[string]sqldb.Client  `json:"-"` // PrepareSQLDBClients()
	services            []svc.Service            // Services to Manage
	done                chan error

	// ToDo: Refactor as fields of each sql db client
	MainDBRawStore     *sqldb.RawStore            `json:"-"`
	MainDBPlaceholder  func(...int) string        `json:"-"`
	MainDBPlaceholders func(int, ...int) []string `json:"-"`
}

// BaseInit - 1st step for initialization
// 1. set AppRoot
// 2. load config/.core.json file
// 3. prepare base fields
// 4. Start ShutdownSignalListener
func (c *Core[B]) BaseInit(appRoot string, rootCtx context.Context, rootCancel context.CancelFunc) error {
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

func (c *Core[B]) prepareDefaultFeatures() {
	c.VolatileKV = &sync.Map{}
	c.SessionLocks = &sync.Map{}
	c.HttpClient = &http.Client{}
	c.ActionLocks = &sync.Map{}
}

func (c *Core[B]) AddService(s svc.Service) {
	log.Printf("[INFO] adding service: %s", s.Name())
	c.services = append(c.services, s)
	log.Printf("[INFO] total services: %d", len(c.services))
}

func (c *Core[B]) StartServices() error {
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

func (c *Core[B]) WaitServicesDone() error {
	for i := 0; i < len(c.services); i++ {
		if err := <-c.done; err != nil {
			return err
		}
	}
	return nil
}

func (c *Core[B]) StopServices() {
	for _, s := range c.services {
		s.Stop()
	}
}

var once sync.Once

func (c *Core[B]) startShutdownSignalListener() {
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

func (c *Core[B]) PrepareJobScheduler() {
	c.JobScheduler = schedjobs.NewScheduler(c.RootCtx)
	c.AddService(c.JobScheduler)
}

func (c *Core[B]) PrepareUDSService(sockPath string, cmdMap map[string]uds.CmdHnd) {
	c.UDSService = uds.NewService(c.RootCtx, sockPath, cmdMap)
	c.AddService(c.UDSService)
}

func (c *Core[B]) PrepareWebService(addr string, router http.Handler) {
	c.WebService = web.NewService(c.RootCtx, addr, router)
	c.AddService(c.WebService)
}

func (c *Core[B]) PrepareThrottleBucketStore(cleanupCycle time.Duration, cleanupOlderThan time.Duration) {
	c.ThrottleBucketStore = throttle.NewBucketStore[B](c.RootCtx, cleanupCycle, cleanupOlderThan)
	c.AddService(c.ThrottleBucketStore)
}

func (c *Core[B]) LoadStorageConf() error {
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

func (c *Core[B]) PrepareKVDatabase() error {
	// Load KV Database Config File
	err := c.LoadKVDBConf()
	if err != nil {
		return err
	}
	if err = c.PrepareKVDBClient(); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) LoadKVDBConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", "databases", ".kv.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.KVDBConf); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) PrepareKVDBClient() error {
	switch c.KVDBConf.Type {
	case "redis":
		c.KVDBClient = &redis.Client{Conf: &c.KVDBConf}
		if err := c.KVDBClient.Init(); err != nil {
			return err
		}
	// case "memcached"
	default:
		return errors.New("unsupported key-value database type")
	}
	return nil
}

func (c *Core[B]) PrepareSQLDatabases(preload func()) error {
	// Load SQL Databases Config File
	err := c.LoadSQLDBConfs()
	if err != nil {
		return err
	}

	// Prepare SQL DB Clients
	if err = c.PrepareSQLDBClients(); err != nil {
		return err
	}

	// ToDo: Refactor this as part of each client
	// Main SQL DB Placeholder Prefix
	mainSQLDBConf, ok := c.SQLDBConfs["main"]
	if !ok {
		return errors.New("main SQL DB conf not found")
	}
	placeholderPrefix, ok := sqldb.PlaceholderPrefixForDBType[mainSQLDBConf.Type]
	if !ok {
		return errors.New("unsupported database type: " + mainSQLDBConf.Type)
	}
	// Main SQL DB Placeholder Gen Fns
	c.MainDBPlaceholder = sqldb.PlaceholderGF(placeholderPrefix)
	c.MainDBPlaceholders = sqldb.PlaceholdersGF(placeholderPrefix)
	// Main SQL DB SQL RawStore
	c.PrepareMainDBRawStore()
	// Preload SQL Models & Packages
	if preload != nil {
		preload()
	}
	err = sqldb.LoadRawStmtsToStore(c.MainDBRawStore, mainSQLDBConf.Type, placeholderPrefix)
	if err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) LoadSQLDBConfs() error {
	confFilePath := filepath.Join(c.AppRoot, "config", "databases", ".sql.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	c.SQLDBConfs = make(map[string]*sqldb.Conf)
	if err = json.Unmarshal(confBytes, &c.SQLDBConfs); err != nil {
		return err
	}
	return nil
}

// PrepareSQLDBClients - Build & Init SQL DB Clients
// Use after LoadSQLDBConfs
func (c *Core[B]) PrepareSQLDBClients() error {
	c.SQLDBClients = make(map[string]sqldb.Client)
	// Registering Supported Implementations
	pgsql.Register()
	mysql.Register()
	// Prepare New Clients
	for dbName, sqlDBConf := range c.SQLDBConfs {
		dbClient, err := sqldb.New(sqlDBConf.Type, sqlDBConf)
		if err != nil {
			return err
		}
		if err = dbClient.Init(); err != nil {
			return err
		}
		// ToDo: Raw SQL Store
		// ToDo: Placeholder Generating Functions
		c.SQLDBClients[dbName] = dbClient
	}
	return nil
}

func (c *Core[B]) PrepareMainDBRawStore() {
	c.MainDBRawStore = sqldb.NewRawStore()
}

func (c *Core[B]) ResourceCleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")
	// Clean up DB clients ----
	// ToDo: factor out this
	if c.KVDBClient != nil {
		if err := c.KVDBClient.Close(); err != nil {
			log.Println("[ERROR] Failed to close KV database client")
		}
	}
	for name, sqlDBClient := range c.SQLDBClients {
		dbType := sqlDBClient.GetConf().Type
		log.Printf("[INFO][%s] Closing %q SQL DB client", dbType, name)
		err := sqlDBClient.Close()
		if err != nil {
			log.Printf("[ERROR][%s] Failed to close %q SQL DB client", dbType, name)
		} else {
			log.Printf("[INFO][%s] %q SQL DB client closed", dbType, name)
		}
	}
	//----
	log.Println("[INFO] App Resource Cleanup Complete")
}

type DebugOpts struct {
	MaintenanceMode int `json:"maintenance_mode"`
	AuthBreak       int `json:"auth_break"`
}
