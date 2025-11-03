package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/zeptools/gw-core/db/sqldb"

	_ "github.com/go-sql-driver/mysql" // side-effect
)

var rawStore = sqldb.NewRawStore()

type Client struct {
	Handle // [Embedded] for Promoted Methods
	conf   *sqldb.Conf
	dsn    string
}

// Ensure mysql.Client implements sqldb.Client interface
var _ sqldb.Client = (*Client)(nil)

func NewClient(conf *sqldb.Conf) (sqldb.Client, error) {
	return &Client{conf: conf}, nil
}

func Register() {
	sqldb.RegisterFactory("mysql", NewClient)
}

func (c *Client) Init() error {
	if c.conf.DSN != "" {
		c.dsn = c.conf.DSN
	} else {
		c.dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=%s&sql_mode=ANSI_QUOTES&multiStatements=true",
			c.conf.User,
			c.conf.PW,
			c.conf.Host,
			c.conf.Port,
			c.conf.DB,
			c.conf.TZ,
		)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Open
	err := c.Open(ctx)
	if err != nil {
		return err
	}
	// Ping
	if err = c.Ping(ctx); err != nil {
		return err
	}
	log.Println("[INFO] mysql client initialized")
	return nil
}

func (c *Client) DBHandle() sqldb.Handle {
	return &Handle{DB: c.DB}
}

func (c *Client) Conf() *sqldb.Conf {
	return c.conf
}

func (c *Client) DSN() string {
	return c.dsn
}

func (c *Client) RawSQLStore() *sqldb.RawStore {
	return rawStore
}

func (c *Client) Open(_ context.Context) error {
	var err error
	if c.DB, err = sql.Open("mysql", c.dsn); err != nil {
		return err
	}
	// ToDo: get this values from conf
	c.SetConnMaxLifetime(time.Minute * 3)
	c.SetMaxOpenConns(10)
	c.SetMaxIdleConns(10)
	return nil
}

func (c *Client) Close() error {
	if c.DB == nil {
		return nil
	}
	log.Println("[INFO] closing mysql client")
	err := c.DB.Close()
	if err != nil {
		return err
	}
	log.Println("[INFO] mysql client closed")
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.PingContext(ctx)
}

func (c *Client) BeginTx(ctx context.Context) (sqldb.Tx, error) {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}
