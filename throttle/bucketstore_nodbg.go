//go:build !debug

package throttle

import (
	"time"
)

func (s *BucketStore[K]) Cleanup(now time.Time) {
	for _, g := range s.groups {
		g.buckets.Range(func(id, value any) bool {
			b := value.(*Bucket[K])
			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			b.mu.Unlock()
			if now.Sub(last) > s.cleanupOlderThan {
				g.buckets.Delete(id)
			}
			return true // continue iteration
		})
	}
}
