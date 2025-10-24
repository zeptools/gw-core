//go:build debug

package throttle

import (
	"log"
	"time"
)

func (s *BucketStore[K]) Cleanup(olderThan time.Duration, now time.Time) {
	log.Printf("[DEBUG] cleaning expired buckets older than %v at %v", olderThan, now)
	cleanCnt := 0
	for gid, g := range s.groups {
		log.Printf("[DEBUG] cleaning expired buckets in bucketgroup %q", gid)
		g.buckets.Range(func(id, value any) bool {
			b := value.(*Bucket[K])

			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			log.Printf("[DEBUG] inspecting bucket %q lastCheck = %v", id, last)
			b.mu.Unlock()

			if now.Sub(last) > olderThan {
				g.buckets.Delete(id)
				cleanCnt++
				log.Printf("[DEBUG] expired bucket %q removed", id)
			}
			return true // continue iteration
		})
	}
	log.Printf("[DEBUG] %d buckets cleaned up", cleanCnt)
}
