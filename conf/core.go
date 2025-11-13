package conf

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/zeptools/gw-core/apis/mainbackend"
	"github.com/zeptools/gw-core/clients"
	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/db/kvdb/impls/redis"
	"github.com/zeptools/gw-core/db/sqldb"
	"github.com/zeptools/gw-core/db/sqldb/impls/mysql"
	"github.com/zeptools/gw-core/db/sqldb/impls/pgsql"
	"github.com/zeptools/gw-core/schedjobs"
	"github.com/zeptools/gw-core/security"
	"github.com/zeptools/gw-core/storages"
	"github.com/zeptools/gw-core/svc"
	"github.com/zeptools/gw-core/throttle"
	"github.com/zeptools/gw-core/tpl"
	"github.com/zeptools/gw-core/uds"
	"github.com/zeptools/gw-core/web"
	"github.com/zeptools/gw-core/web/session"
)

// Core - common config
// B = Throttle BucketID Type _ e.g. string, int64, etc
type Core[B comparable] struct {
	AppName             string                                           `json:"app_name"`
	Listen              string                                           `json:"listen"`     // HTTP Server Listen IP:PORT Address
	Host                string                                           `json:"host"`       // HTTP Host. Can be used to generate public url endpoints
	DebugOpts           DebugOpts                                        `json:"debug_opts"` // Debug Options
	AppRoot             string                                           `json:"-"`          // Filled from compiled paths
	RootCtx             context.Context                                  `json:"-"`          // Global Context with RootCancel
	RootCancel          context.CancelFunc                               `json:"-"`          // CancelFunc for RootCtx
	UDSService          *uds.Service                                     `json:"-"`          // PrepareUDSService
	JobScheduler        *schedjobs.Scheduler                             `json:"-"`          // PrepareJobScheduler
	WebService          *web.Service                                     `json:"-"`          // PrepareWebService
	ThrottleBucketStore *throttle.BucketStore[B]                         `json:"-"`          // PrepareThrottleBucketStore
	VolatileKV          *sync.Map                                        `json:"-"`          // map[string]string
	SessionLocks        *sync.Map                                        `json:"-"`          // map[string]*sync.Mutex for ServiceSessions and WebSessions
	ActionLocks         *sync.Map                                        `json:"-"`          // map[string]struct{}
	StorageConf         storages.Conf                                    `json:"-"`          // LoadStorageConf
	BackendHttpClient   *http.Client                                     `json:"-"`          // for requests to external apis
	KVDBConf            kvdb.Conf                                        `json:"-"`          // loadKVDBConf
	BackendKVDBClient   kvdb.Client                                      `json:"-"`          // prepareKVDBClient
	SQLDBConfs          map[string]*sqldb.Conf                           `json:"-"`          // loadSQLDBConfs
	BackendSQLDBClients map[string]sqldb.Client                          `json:"-"`          // prepareSQLDBClients
	ClientApps          atomic.Pointer[map[string]clients.ClientAppConf] `json:"-"`          // [Hot Reload] PrepareClientApps
	WebSessionManager   *session.Manager                                 `json:"-"`          // PrepareWebSessions
	MainBackendClient   *mainbackend.Client                              `json:"-"`          // PrepareMainBackendClient
	HTMLTemplateStore   *tpl.HTMLTemplateStore                           `json:"-"`          // PrepareHTMLTemplateStore

	services []svc.Service // Services to Manage
	done     chan error
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
	c.BackendHttpClient = &http.Client{}
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

func (c *Core[B]) PrepareUDSService(sockPath string, cmdStore *uds.CommandStore) {
	c.UDSService = uds.NewService(c.RootCtx, sockPath, cmdStore)
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
	err := c.loadKVDBConf()
	if err != nil {
		return err
	}
	if err = c.prepareKVDBClient(); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) loadKVDBConf() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".kv-databases.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(confBytes, &c.KVDBConf); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) prepareKVDBClient() error {
	switch c.KVDBConf.Type {
	case "redis":
		c.BackendKVDBClient = &redis.Client{Conf: &c.KVDBConf}
		if err := c.BackendKVDBClient.Init(); err != nil {
			return err
		}
	// case "memcached"
	default:
		return errors.New("unsupported key-value database type")
	}
	return nil
}

func (c *Core[B]) loadSQLDBConfs() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".sql-databases.json")
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

// prepareSQLDBClients - Build & Init SQL DB Clients
// Use after loadSQLDBConfs
func (c *Core[B]) prepareSQLDBClients() error {
	c.BackendSQLDBClients = make(map[string]sqldb.Client)

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
		c.BackendSQLDBClients[dbName] = dbClient
	}
	return nil
}

// PrepareSQLDatabases for SQL DB Clients & RawSQL Stores, etc
func (c *Core[B]) PrepareSQLDatabases(ensureImports func()) error {
	// Load SQL Databases Config File
	err := c.loadSQLDBConfs()
	if err != nil {
		return err
	}
	DBTypesSet := make(map[string]struct{})
	for _, conf := range c.SQLDBConfs {
		DBTypesSet[conf.Type] = struct{}{}
	}
	if len(DBTypesSet) == 0 {
		return nil
	}

	// Prepare SQL DB Clients
	if err = c.prepareSQLDBClients(); err != nil {
		return err
	}

	// Load Raw Statements to Stores
	if ensureImports != nil {
		ensureImports()
	}
	if _, ok := DBTypesSet["mysql"]; ok {
		err = mysql.LoadRawStmtsToStore()
		if err != nil {
			return err
		}
	}
	if _, ok := DBTypesSet["pgsql"]; ok {
		err = pgsql.LoadRawStmtsToStore()
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareClientApps prepares ClientApps
// building a new clients.ClientAppConf map and swaps the atomic pointer for the ClientApps
// So, this can be invoked to Hot-Reload the ClientApps
func (c *Core[B]) PrepareClientApps() error {
	var (
		err           error
		newClientApps map[string]clients.ClientAppConf
	)
	if newClientApps, err = c.newClientAppsConfMapFromFile(); err != nil {
		return err
	}
	c.ClientApps.Store(&newClientApps) // atomic store
	return nil
}

func (c *Core[B]) newClientAppsConfMapFromFile() (map[string]clients.ClientAppConf, error) {
	confFilePath := filepath.Join(c.AppRoot, "config", ".clients.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return nil, err
	}
	var confMap map[string]clients.ClientAppConf
	if err = json.Unmarshal(confBytes, &confMap); err != nil {
		return nil, err
	}
	return confMap, nil
}

// GetClientAppConf reads a clients.ClientAppConf
// Uses a single atomic cpu instruction
func (c *Core[B]) GetClientAppConf(id string) (clients.ClientAppConf, bool) {
	confMapPtr := c.ClientApps.Load()
	if confMapPtr == nil {
		return clients.ClientAppConf{}, false
	}
	conf, ok := (*confMapPtr)[id]
	conf.ID = id
	return conf, ok
}

// PrepareWebSessions prepares WebSessionManager
// Prerequisite: BackendKVDBClient
// Prerequisite: SessionLocks
func (c *Core[B]) PrepareWebSessions() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".web-session.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if c.BackendKVDBClient == nil {
		return errors.New("backend KVDB client not ready")
	}
	if c.SessionLocks == nil {
		return errors.New("sessionlocks not ready")
	}
	mgr := &session.Manager{
		AppName:           c.AppName,
		BackendKVDBClient: c.BackendKVDBClient,
		SessionLocks:      c.SessionLocks,
	}
	if err = json.Unmarshal(confBytes, &mgr.Conf); err != nil {
		return err
	}
	// Web Login Session Cipher
	cipher, err := security.NewXChaCha20Poly1305CipherBase64([]byte(mgr.Conf.EncryptionKey))
	if err != nil {
		return fmt.Errorf("NewXChaCha20Poly1305Cipher: %v", err)
	}
	mgr.Cipher = cipher

	c.WebSessionManager = mgr
	return nil
}

// PrepareMainBackendClient to Send Request to the Main Backend API if any
// Prerequisite: BackendHttpClient
func (c *Core[B]) PrepareMainBackendClient() error {
	confFilePath := filepath.Join(c.AppRoot, "config", ".main-backend-api.json")
	confBytes, err := os.ReadFile(confFilePath) // ([]byte, error)
	if err != nil {
		return err
	}
	if c.BackendHttpClient == nil {
		return errors.New("backend http client not ready")
	}
	c.MainBackendClient = &mainbackend.Client{
		Client: c.BackendHttpClient,
	}
	if err = json.Unmarshal(confBytes, &c.MainBackendClient.Conf); err != nil {
		return err
	}
	return nil
}

func (c *Core[B]) PrepareHTMLTemplateStore() error {
	c.HTMLTemplateStore = tpl.NewHTMLTemplateStore()
	return c.HTMLTemplateStore.LoadBaseTemplates(
		filepath.Join(c.AppRoot, "templates", "html"),
	)
}

func (c *Core[B]) ResourceCleanUp() {
	log.Println("[INFO] App Resource Cleaning Up...")
	// Clean up DB clients ----
	// ToDo: factor out this
	if c.BackendKVDBClient != nil {
		if err := c.BackendKVDBClient.Close(); err != nil {
			log.Println("[ERROR] Failed to close KV database client")
		}
	}
	for name, sqlDBClient := range c.BackendSQLDBClients {
		dbType := sqlDBClient.Conf().Type
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
