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
		log.Printf("[DEBUG] cleaning up bucketgroup %q", gid)
		g.buckets.Range(func(id, value any) bool {
			b := value.(*Bucket[K])

			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			switch typedID := id.(type) {
			case int64, int:
				log.Printf("[DEBUG] inspecting bucket id=%d lastCheck=%v", typedID, last)
			case string:
				log.Printf("[DEBUG] inspecting bucket id=%q lastCheck=%v", typedID, last)
			default:
				log.Printf("[DEBUG] inspecting bucket id=%v lastCheck=%v", typedID, last)
			}
			b.mu.Unlock()

			if now.Sub(last) > olderThan {
				g.buckets.Delete(id)
				cleanCnt++
				log.Println("[DEBUG] bucket REMOVED")
			} else {
				log.Println("[DEBUG] keeping bucket")
			}
			return true // continue iteration
		})
	}
	log.Printf("[DEBUG] %d buckets cleaned up", cleanCnt)
}
