package state

import (
	"errors"
	"log"
	"strconv"
	"time"
)

var (
	ErrorNotInteger = errors.New("ERR value is not an integer or out of range")
)

type DbString struct {
	string
	expiresAt time.Time
}

var _ DbValue = (*DbString)(nil)
var _ DefinitelyExpirer = (*DbString)(nil)

func (v *DbString) Type() string {
	return "string"
}

func (v *DbString) IsDefinitelyExpiredAt(t time.Time) bool {
	return !IsReplica() && !v.expiresAt.IsZero() && v.expiresAt.Before(t)
}

func Set(key, value string, px int64) error {
	// zero time means no expiry
	var expiresAt time.Time
	if px != -1 {
		expiresAt = time.Now().Add(time.Duration(px) * time.Millisecond)
	}
	state.DbMu.Lock()
	UnsafeSet(key, value, expiresAt)
	state.DbMu.Unlock()
	return nil
}

func Incr(key string) (int64, error) {
	state.DbMu.Lock()
	defer state.DbMu.Unlock()
	v, ok := state.Db[key]
	if !ok {
		state.Db[key] = &DbString{string: "1", expiresAt: time.Time{}}
		return 1, nil
	}
	w, ok := v.(*DbString)
	if !ok {
		return 0, ErrorNotInteger
	}
	if !w.expiresAt.IsZero() && w.expiresAt.Before(time.Now()) {
		state.Db[key] = &DbString{string: "1", expiresAt: time.Time{}}
		return 1, nil
	}
	i, err := strconv.ParseInt(w.string, 10, 64)
	if err != nil {
		return 0, ErrorNotInteger
	}
	i++
	UnsafeSet(key, strconv.FormatInt(i, 10), w.expiresAt)
	return i, nil
}

func Get(key string) (string, error) {
	state.DbMu.RLock()
	v, ok := state.Db[key]
	state.DbMu.RUnlock()
	if !ok {
		return "", ErrorNone
	}
	w, ok := v.(*DbString)
	if !ok {
		return "", ErrorWrongType
	}
	if !w.expiresAt.IsZero() && w.expiresAt.Before(time.Now()) {
		go TryEvictExpiredKey(key)
		return "", ErrorNone
	}
	return w.string, nil
}

func TryEvictExpiredKey(key string) {
	state.DbMu.Lock()
	defer state.DbMu.Unlock()
	v, ok := state.Db[key]
	if !ok {
		return
	}
	if w, ok := v.(DefinitelyExpirer); ok && w.IsDefinitelyExpiredAt(time.Now()) {
		delete(state.Db, key)
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

	if IsReplica() {
		log.Println("SyncTryEvictExpiredKeysSweep: not a master, skipping")
		return
	}
	log.Println("SyncTryEvictExpiredKeysSweep: started")
	state.DbMu.RLock()
	if len(state.Db) < evictionSweepMapSizeThreshold {
		return
	}
	state.DbMu.RUnlock()
	now := time.Now()
	state.DbMu.Lock()
	i := 0
	for k, v := range state.Db {
		if w, ok := v.(DefinitelyExpirer); ok && w.IsDefinitelyExpiredAt(now) {
			delete(state.Db, k)
		}
		i++
		if i >= evictionSweepCountPerAcquisition {
			state.DbMu.Unlock()
			i = 0
			time.Sleep(evictionSweepSleepPerAcquisition)
			now = time.Now()
			state.DbMu.Lock()
		}
	}
	state.DbMu.Unlock()
}
