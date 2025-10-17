package keyonlylocks

import "sync"

func AcquireLocks(lockStore *sync.Map, keys []string) ([]string, bool) {
	//sort.Strings(keys) // optional, prevents deadlocks if WAIT mode added
	var acquired []string
	for _, key := range keys {
		_, loaded := lockStore.LoadOrStore(key, struct{}{})
		if loaded {
			// rollback previously acquired locks
			for _, k := range acquired {
				lockStore.Delete(k)
			}
			return nil, false
		}
		acquired = append(acquired, key)
	}
	return acquired, true
}

// ReleaseLocks delete locks from the lockStore *sync.Map
// Wrap this in deferred calls to guarantee to be called even if panic occurs.
func ReleaseLocks(lockStore *sync.Map, keys []string) {
	for _, key := range keys {
		lockStore.Delete(key)
	}
}
