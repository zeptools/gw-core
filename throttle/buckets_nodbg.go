//go:build !debug

package throttle

import (
	"time"
)

func (s *BucketStore[K]) Cleanup(olderThan time.Duration, now time.Time) {
	for _, g := range s.groups {
		g.buckets.Range(func(key, value any) bool {
			b := value.(*Bucket[K])
			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			b.mu.Unlock()

			if now.Sub(last) > olderThan {
				g.buckets.Delete(key)
			}
			return true // continue iteration
		})
	}
}
