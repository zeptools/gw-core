package throttle

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zeptools/gw-core/svc"
)

type BucketStore[K comparable] struct {
	Ctx              context.Context    // Service Context
	cancel           context.CancelFunc // Service Context CancelFunc
	state            int                // internal service state
	done             chan error         // Shutdown Error Channel
	cleanupCycle     time.Duration
	cleanupOlderThan time.Duration
	groups           map[string]*BucketGroup[K]
}

func (s *BucketStore[K]) Name() string {
	return "ThrottleBucketStore"
}

func NewBucketStore[K comparable](parentCtx context.Context, cleanupCycle time.Duration, cleanupOlderThan time.Duration) *BucketStore[K] {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &BucketStore[K]{
		Ctx:              svcCtx,
		cancel:           svcCancel,
		state:            svc.StateREADY,
		cleanupCycle:     cleanupCycle,
		cleanupOlderThan: cleanupOlderThan,
		groups:           make(map[string]*BucketGroup[K]),
	}
}

// Start starts a service that manages buckets
func (s *BucketStore[K]) Start() error {
	if s.state == svc.StateRUNNING {
		return fmt.Errorf("already started")
	}
	if s.state != svc.StateREADY {
		return fmt.Errorf("cannot start. not ready")
	}
	s.state = svc.StateRUNNING
	log.Printf("[INFO][Throttle] cleanup service started cycle=%v exp=%v", s.cleanupCycle, s.cleanupOlderThan)
	go s.run()
	return nil
}

func (s *BucketStore[K]) Stop() {
	if s.state != svc.StateRUNNING {
		log.Println("[ERROR][Throttle] cannot stop. not running")
		return
	}
	s.cancel()
	s.state = svc.StateSTOPPED
	log.Println("[INFO][Throttle] service stopped")
}

func (s *BucketStore[K]) Done() <-chan error {
	return s.done
}

func (s *BucketStore[K]) run() {
	ticker := time.NewTicker(s.cleanupCycle)
	defer ticker.Stop()
	for {
		select {
		case <-s.Ctx.Done():
			log.Println("[INFO][Throttle] stopping cleaning service")
			s.done <- nil
			return
		case now := <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PANIC] recovered in throttle bucketstore cleaning service: %v", r)
					}
				}()
				log.Printf("[INFO][Throttle] %v cleanup cycle ...", s.cleanupCycle)
				s.Cleanup(now)
			}()
		}
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
