package throttle

import (
	"sync"
	"time"
)

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
