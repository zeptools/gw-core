//go:build debug

package throttle

import (
	"log"
	"time"
)

func (s *BucketStore[K]) Cleanup(olderThan time.Duration, now time.Time) {
	log.Printf("[DEBUG] cleaning expired buckets older than %v", olderThan)
	for gid, g := range s.groups {
		g.buckets.Range(func(key, value any) bool {
			b := value.(*Bucket[K])
			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			b.mu.Unlock()

			if now.Sub(last) > olderThan {
				g.buckets.Delete(key)
				log.Printf("[DEBUG] expired bucket %q removed in %q", key, gid)
			}
			return true // continue iteration
		})
	}
	log.Printf("[DEBUG] expired buckets cleaned up")
}
