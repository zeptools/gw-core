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

type Client struct {
	//sqldb.Client // [Embedded Interface]

	Conf *sqldb.Conf

	// db fields are implementation details, not exported
	db  *sql.DB
	dsn string
}

// Ensure mysql.Client implements sqldb.Client interface
var _ sqldb.Client = (*Client)(nil)

func (c *Client) Init() error {
	var err error
	if c.Conf.DSN != "" {
		c.dsn = c.Conf.DSN
	} else {
		c.dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=%s&sql_mode=ANSI_QUOTES&multiStatements=true",
			c.Conf.User,
			c.Conf.PW,
			c.Conf.Host,
			c.Conf.Port,
			c.Conf.DB,
			c.Conf.TZ,
		)
	}
	if c.db, err = sql.Open("mysql", c.dsn); err != nil {
		return err
	}
	c.db.SetConnMaxLifetime(time.Minute * 3)
	c.db.SetMaxOpenConns(10)
	c.db.SetMaxIdleConns(10)
	if err = c.db.Ping(); err != nil {
		return err
	}
	log.Println("[INFO] mysql client initialized")
	return nil
}

func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	log.Println("[INFO] closing mysql client")
	err := c.db.Close()
	if err != nil {
		return err
	}
	log.Println("[INFO] mysql client closed")
	return nil
}

func (c *Client) DBHandle() sqldb.DBHandle {
	return &DBHandle{db: c.db}
}

func (c *Client) BeginTx(ctx context.Context) (sqldb.Tx, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}
