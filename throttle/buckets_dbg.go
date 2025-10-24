//go:build debug

package throttle

import (
	"log"
	"time"
)

func (s *BucketStore[K]) Cleanup(olderThan time.Duration, now time.Time) {
	log.Printf("[DEBUG][Throttle] cleaning expired Buckets older than %v at %v", olderThan, now)
	cleanCnt := 0
	for gid, g := range s.groups {
		log.Printf("[DEBUG][Throttle] cleaning BucketGroup %q", gid)
		g.buckets.Range(func(id, value any) bool {
			b := value.(*Bucket[K])

			// lock per bucket while checking/removing
			b.mu.Lock()
			last := b.lastCheck
			switch typedID := id.(type) {
			case int64, int:
				log.Printf("[DEBUG][Throttle] inspecting Bucket id=%d lastCheck=%v", typedID, last)
			case string:
				log.Printf("[DEBUG][Throttle] inspecting Bucket id=%q lastCheck=%v", typedID, last)
			default:
				log.Printf("[DEBUG][Throttle] inspecting Bucket id=%v lastCheck=%v", typedID, last)
			}
			b.mu.Unlock()

			if now.Sub(last) > olderThan {
				g.buckets.Delete(id)
				cleanCnt++
				log.Println("[DEBUG][Throttle] Bucket REMOVED")
			} else {
				log.Println("[DEBUG][Throttle] keeping Bucket")
			}
			return true // continue iteration
		})
	}
	log.Printf("[DEBUG][Throttle] %d Buckets cleaned up", cleanCnt)
}
