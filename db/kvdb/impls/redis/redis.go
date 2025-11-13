package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zeptools/gw-core/db/kvdb"

	lowimpl "github.com/redis/go-redis/v9"
)

type Client struct {
	Conf *kvdb.Conf

	// implementation details, not exported
	internal *lowimpl.Client
}

// Ensure redis.Client implements kvdb.Client interface
var _ kvdb.Client = (*Client)(nil)

func (c *Client) Init() error {
	c.internal = lowimpl.NewClient(&lowimpl.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Conf.Host, c.Conf.Port),
		Password: c.Conf.PW,
		DB:       c.Conf.DB,
	})
	log.Println("[INFO] redis internal initialized")
	return nil
}

func (c *Client) Close() error {
	if c.internal == nil {
		return nil
	}
	return c.internal.Close()
}

func (c *Client) GetHandle() any { // use with runtime type assertion
	return c.internal
}

func (c *Client) GetConf() *kvdb.Conf {
	return c.Conf
}

//--- Key Ops ----

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.internal.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Client) Delete(ctx context.Context, keys ...string) (int64, error) {
	return c.internal.Del(ctx, keys...).Result()
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// Redis EXPIRE returns true if key existed and TTL was set, false if key does not exist
	return c.internal.Expire(ctx, key, expiration).Result()
}

func (c *Client) Type(ctx context.Context, key string) (string, error) {
	return c.internal.Type(ctx, key).Result()
}

func (c *Client) ScanKeys(ctx context.Context, cursor any, scanBatchSize int) ([]string, any, error) {
	var cur uint64
	if cursor != nil {
		cur = cursor.(uint64)
	}
	keys, nextCursor, err := c.internal.Scan(ctx, cur, "*", int64(scanBatchSize)).Result()
	if err != nil {
		return nil, nil, err
	}
	// Redis returns nextCursor == 0 when the scan is complete.
	if nextCursor == 0 {
		return keys, nil, nil
	}
	return keys, nextCursor, nil
}

//---- Single-value Ops ----

func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.internal.Get(ctx, key).Result()
	if errors.Is(err, lowimpl.Nil) {
		return "", false, nil // redis.Nil -> ok: false, err: nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.internal.Set(ctx, key, value, expiration).Err()
}

//---- List Ops ----

func (c *Client) Push(ctx context.Context, key, value string) error {
	// AddHandlers to the tail (right) of the list
	return c.internal.RPush(ctx, key, value).Err()
}

func (c *Client) Pop(ctx context.Context, key string) (string, bool, error) { // val, found, err
	// Pop from the head (left) of the list (FIFO)
	val, err := c.internal.LPop(ctx, key).Result()
	if errors.Is(err, lowimpl.Nil) {
		return "", false, nil // redis.Nil -> ok: false, err: nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *Client) Len(ctx context.Context, key string) (int64, error) {
	return c.internal.LLen(ctx, key).Result()
}

func (c *Client) Range(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.internal.LRange(ctx, key, start, stop).Result()
}

func (c *Client) Remove(ctx context.Context, key string, cnt int64, value any) (int64, error) {
	return c.internal.LRem(ctx, key, cnt, value).Result()
}

func (c *Client) Trim(ctx context.Context, key string, start, stop int64) error {
	return c.internal.LTrim(ctx, key, start, stop).Err()
}

//---- Hash Ops ----

func (c *Client) SetField(ctx context.Context, key string, field string, value any) error {
	return c.internal.HSet(ctx, key, field, value).Err()
}

func (c *Client) GetField(ctx context.Context, key string, field string) (string, bool, error) { // val, found, err
	val, err := c.internal.HGet(ctx, key, field).Result()
	if errors.Is(err, lowimpl.Nil) {
		return "", false, nil // key or field missing
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *Client) SetFields(ctx context.Context, key string, fields map[string]any) error {
	return c.internal.HSet(ctx, key, fields).Err()
}

// GetFields returns a map {field:value} from a hash data, which contains only found fields
// so, if len(rtnMap) < len(fields), some fields are missing
// [NOTE] returns an empty map even if key is not found. not error
func (c *Client) GetFields(ctx context.Context, key string, fields ...string) (map[string]string, error) {
	resultSlice, err := c.internal.HMGet(ctx, key, fields...).Result() // []any
	if err != nil {
		return nil, err
	}
	rtnMap := make(map[string]string, len(fields)) // capacity = max len = when all fields found
	for i, v := range resultSlice {
		if v != nil {
			rtnMap[fields[i]] = fmt.Sprint(v)
		}
		// if v is nil, field missing → omitted in the return map
	}
	return rtnMap, nil
}

func (c *Client) RemoveFields(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.internal.HDel(ctx, key, fields...).Result()
}

// GetAllFields returns a map {field:value} from a hash data with all fields in it
// [NOTE] returns an empty map even if key is not found. not error
func (c *Client) GetAllFields(ctx context.Context, key string) (map[string]string, error) {
	return c.internal.HGetAll(ctx, key).Result()
}
