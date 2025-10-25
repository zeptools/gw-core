package throttle

import (
	"log"
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

type BucketGroup[K comparable] struct {
	conf    *BucketConf
	buckets *sync.Map // K -> *Bucket[K]
}

func (g *BucketGroup[K]) GetBucket(id K) (*Bucket[K], bool) {
	bAny, ok := g.buckets.Load(id)
	if !ok {
		return nil, false
	}
	return bAny.(*Bucket[K]), true
}

func (g *BucketGroup[K]) SetBucket(id K, tokens int, now time.Time) {
	g.buckets.Store(id, &Bucket[K]{
		tokens:      tokens,
		lastCheck:   now,
		parentGroup: g,
	})
}

type BucketConf struct {
	Burst     int           // maximum number of tokens in the bucket
	Increment int           // how many tokens to add each period
	Period    time.Duration // how often to add Increment
}

type BucketStore[K comparable] struct {
	groups map[string]*BucketGroup[K]
}

func NewBucketStore[K comparable]() *BucketStore[K] {
	return &BucketStore[K]{
		groups: make(map[string]*BucketGroup[K]),
	}
}

func (s *BucketStore[K]) GetBucketGroup(id string) (*BucketGroup[K], bool) {
	g, ok := s.groups[id]
	return g, ok
}

func (s *BucketStore[K]) GetBucket(groupID string, userID K) (*Bucket[K], bool) {
	g, ok := s.groups[groupID]
	if !ok {
		return nil, false
	}
	return g.GetBucket(userID)
}

func (s *BucketStore[K]) SetBucketGroup(id string, conf *BucketConf) {
	s.groups[id] = &BucketGroup[K]{
		conf:    conf,
		buckets: &sync.Map{},
	}
}

func (s *BucketStore[K]) Allow(groupID string, userID K, now time.Time) bool {
	g, ok := s.GetBucketGroup(groupID)
	if !ok {
		return false // Invalid groupID always Blocked
	}
	b, ok := g.GetBucket(userID)
	if ok {
		return b.Allow(now)
	}
	// consume 1 token from the fresh bucket
	g.SetBucket(userID, g.conf.Burst-1, now)
	return true
}

// StartCleanUpService starts a background goroutine that periodically
// cleans up expired buckets. It runs forever until the process exits.
//   - period: how often to wake up
//   - olderThan: how old a bucket must be to be deleted
func (s *BucketStore[K]) StartCleanUpService(period time.Duration, olderThan time.Duration) {
	log.Printf("[INFO][Throttle] starting cleanup service period=%v olderthan=%v", period, olderThan)
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for now := range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC] recovered in throttle bucketstore StartCleanUpService: %v", r)
					}
				}()
				log.Println("[INFO][Throttle] cleanup cycle ...")
				s.Cleanup(olderThan, now)
			}()
		}
	}()
}
