package pgsql

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeptools/gw-core/db/sqldb"
)

type Client struct {
	Handle // [Embedded] for Promoted Methods
	conf   *sqldb.Conf
	dsn    string
}

// Ensure pgsql.Client implements sqldb.Client interface
var _ sqldb.Client = (*Client)(nil)

func NewClient(conf *sqldb.Conf) (sqldb.Client, error) {
	return &Client{conf: conf}, nil
}

func (c *Client) Init() error {
	// DSN
	if c.conf.DSN != "" {
		c.dsn = c.conf.DSN
	} else {
		// NOTE: sslmode=disable is often used for local dev, adjust as needed.
		// NOTE: PostgreSQL natively allows multiple statements in a single query string.
		c.dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
			c.conf.Host,
			c.conf.Port,
			c.conf.User,
			c.conf.PW,
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
		return fmt.Errorf("postgres ping failed: %w", err)
	}
	log.Print("[INFO] pgsql client initialized")
	return nil
}

func (c *Client) DBHandle() sqldb.Handle {
	return &Handle{Pool: c.Pool}
}

func (c *Client) Conf() *sqldb.Conf {
	return c.conf
}

func (c *Client) DSN() string {
	return c.dsn
}

func (c *Client) SinglePlaceholder(nth ...int) string {
	if len(nth) == 0 {
		// No-index Provided
		return DefaultSinglePlaceholder
	}
	return fmt.Sprintf("%c%d", DefaultPlaceholderPrefix, nth[0])
}

func (c *Client) Placeholders(cnt int, start ...int) string {
	placeholders := make([]string, cnt)
	var startI int
	if len(start) == 0 {
		startI = 1
	} else {
		startI = start[0]
	}
	j := startI
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("%c%d", DefaultPlaceholderPrefix, j)
		j++
	}
	return strings.Join(placeholders, ",")
}

func (c *Client) RawSQLStore() *sqldb.RawSQLStore {
	return rawStmtStore
}

func (c *Client) Open(ctx context.Context) error {
	config, err := pgxpool.ParseConfig(c.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse pgx config: %w", err)
	}
	// Pool tuning _ ToDo: get this values from conf
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
