package kvdb

import (
	"context"
	"time"

	"github.com/zeptools/gw-core/db"
)

type Client interface {
	db.Client[any] // Client[T] locked into Client[Any] -> DBHandle() any (Runtime Type Assertion with overhead)

	//---- Key Ops ----

	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, keys ...string) (int64, error)
	// Expire sets/updates expiration for a key
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) // found & updated, err

	//---- Single-value Ops ----

	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, bool, error) // val, found, err

	//---- List Ops ----

	Push(ctx context.Context, key string, value string) error
	Pop(ctx context.Context, key string) (string, bool, error) // val, found, err
	Len(ctx context.Context, key string) (int64, error)
	Range(ctx context.Context, key string, start int64, stop int64) ([]string, error) // 0-basis, stop inclusive
	Remove(ctx context.Context, key string, cnt int64, value any) (int64, error)      // cnt = removed dups. 0 = all
	Trim(ctx context.Context, key string, start int64, stop int64) error              // 0-basis, stop inclusive

	//---- Hash Ops ----

	SetField(ctx context.Context, key string, field string, value any) error
	GetField(ctx context.Context, key string, field string) (string, bool, error) // val, found, err
	SetFields(ctx context.Context, key string, fields map[string]any) error
	// GetFields returns values of found fields. By comparing lengths, you can check if all fields are found
	GetFields(ctx context.Context, key string, fields ...string) (map[string]string, error)
	// RemoveFields removes the specified fields in a hash key. Returns the number of fields actually removed.
	RemoveFields(ctx context.Context, key string, fields ...string) (int64, error)
	GetAllFields(ctx context.Context, key string) (map[string]string, error)
}
