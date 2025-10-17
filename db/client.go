package db

import "log"

type Client[T any] interface {
	Init() error
	Close() error
	DBHandle() T // generic handle
}

func CloseClient[T any](name string, c Client[T]) {
	if c == nil {
		log.Printf("[INFO] `%s` Nothing to Close", name)
		return
	}
	if err := c.Close(); err != nil {
		log.Printf("[WARN] Failed to Close `%s`: %v", name, err)
	} else {
		log.Printf("[INFO] `%s` Closed", name)
	}
}
