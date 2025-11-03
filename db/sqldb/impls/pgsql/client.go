package pgsql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeptools/gw-core/db/sqldb"
)

type Client struct {
	Handle   // [Embedded] for Promoted Methods
	Conf     *sqldb.Conf
	RawStore *sqldb.RawStore
	dsn      string
}

// Ensure pgsql.Client implements sqldb.Client interface
var _ sqldb.Client = (*Client)(nil)

func (c *Client) Init() error {
	// DSN
	if c.Conf.DSN != "" {
		c.dsn = c.Conf.DSN
	} else {
		// NOTE: sslmode=disable is often used for local dev, adjust as needed.
		// NOTE: PostgreSQL natively allows multiple statements in a single query string.
		c.dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
			c.Conf.Host,
			c.Conf.Port,
			c.Conf.User,
			c.Conf.PW,
			c.Conf.DB,
			c.Conf.TZ,
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
		return fmt.Errorf("postgres ping failed: %w", err)
	}
	log.Print("[INFO] pgsql client initialized")
	return nil
}

func (c *Client) GetHandle() sqldb.Handle {
	return &Handle{Pool: c.Pool}
}

func (c *Client) GetConf() *sqldb.Conf {
	return c.Conf
}

func (c *Client) GetDSN() string {
	return c.dsn
}

func (c *Client) GetRawSQLStore() *sqldb.RawStore {
	return c.RawStore
}

func (c *Client) Open(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(c.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse pgx config: %w", err)
	}
	// Pool tuning _ ToDo: get this values from Conf
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 3 * time.Minute
	c.Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect pgx Pool: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	if c.Pool == nil {
		return nil
	}
	log.Println("[INFO] closing pgsql client")
	c.Pool.Close()
	log.Println("[INFO] pgsql client closed")
	return nil
}

func (c *Client) BeginTx(ctx context.Context) (sqldb.Tx, error) {
	if c.Pool == nil {
		return nil, fmt.Errorf("pgsql client not initialized")
	}
	conn, err := c.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection failed: %w", err)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	return &Tx{tx: tx}, nil
}
