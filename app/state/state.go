package state

import (
	"log"
	"sync"
	"time"
)

type stateValue struct {
	string
	expiresAt *time.Time
}

type stateShard struct {
	data map[string]stateValue
	mu   sync.RWMutex
}

var state = stateShard{data: make(map[string]stateValue)}

// getStateShardForKey for now returns a global state variable.
// This function is intended to allow us to shard the state in the future.
// to achieve greater concurrency.
func getStateShardForKey(_ string) *stateShard {
	return &state
}

func getAllMapsWithMutex() []*stateShard {
	return []*stateShard{&state}
}

func Set(key, value string, px int64) {
	ml := getStateShardForKey(key)
	var expiresAt *time.Time = nil
	if px != -1 {
		t := time.Now().Add(time.Duration(px) * time.Millisecond)
		expiresAt = &t
	}
	ml.mu.Lock()
	ml.data[key] = stateValue{string: value, expiresAt: expiresAt}
	ml.mu.Unlock()
}

func Get(key string) (string, bool) {
	ml := getStateShardForKey(key)
	ml.mu.RLock()
	v, ok := ml.data[key]
	ml.mu.RUnlock()
	if !ok {
		return "", false
	}
	if v.expiresAt != nil && v.expiresAt.Before(time.Now()) {
		go tryEvictExpiredKey(key)
		return "", false
	}
	return v.string, true
}

func tryEvictExpiredKey(key string) {
	ml := getStateShardForKey(key)
	ml.mu.Lock()
	defer ml.mu.Unlock()
	v, ok := ml.data[key]
	if !ok {
		return
	}
	if v.expiresAt != nil && v.expiresAt.Before(time.Now()) {
		delete(ml.data, key)
	}
}

// syncTryEvictExpiredKeys is a helper function for daemons to evict expired keys from all maps. It is expected to run for a long time.
func SyncTryEvictExpiredKeysSweep() {
	const (
		// evictionSweepMapSizeThreshold is the number of keys in a map below which we will not bother to sweep for expired keys.
		evictionSweepMapSizeThreshold = 1000
		// evictionSweepCountPerAcquisition is the number of keys we check for expiration each time we acquire the lock
		evictionSweepCountPerAcquisition = 100
		// evictionSweepSleepPerAcquisitionInMs is the number of milliseconds we sleep after each acquisition of the lock.
		evictionSweepSleepPerAcquisition = 10 * time.Millisecond
	)

	log.Println("SyncTryEvictExpiredKeysSweep: started")
	for _, ml := range getAllMapsWithMutex() {
		if len(ml.data) < evictionSweepMapSizeThreshold {
			continue
		}
		now := time.Now()
		ml.mu.Lock()
		i := 0
		for k, v := range ml.data {
			if v.expiresAt != nil && v.expiresAt.Before(now) {
				delete(ml.data, k)
			}
			i++
			if i >= evictionSweepCountPerAcquisition {
				i = 0
				ml.mu.Unlock()
				time.Sleep(evictionSweepSleepPerAcquisition)
				now = time.Now()
				ml.mu.Lock()
			}
		}
		ml.mu.Unlock()
	}
}
