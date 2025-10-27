package throttle

import (
	"sync"
	"time"
)

type Bucket[K comparable] struct {
	mu          sync.Mutex // protects access to bucket state
	tokens      int
	lastCheck   time.Time
	parentGroup *BucketGroup[K] // back-reference to its parentGroup group
}

// refill tokens
// Since this modifies the bucket's state, this should be wrapped by mutex lock/unlock
func (b *Bucket[K]) refill(now time.Time) {
	conf := b.parentGroup.conf
	elapsed := now.Sub(b.lastCheck)
	if elapsed >= conf.Period { // compare
		times := int(elapsed / conf.Period) // division
		b.tokens += times * conf.Increment
		if b.tokens > conf.Burst {
			b.tokens = conf.Burst
		}
		b.lastCheck = b.lastCheck.Add(time.Duration(times) * conf.Period)
	}
}

func (b *Bucket[K]) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(now)
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}
