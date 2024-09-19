package state

import (
	"bufio"
	"encoding/hex"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/rdbreader"
)

var initialized = false

type StateValue struct {
	string
	expiresAt time.Time
}

func NewStateValue(value string, expiresAt time.Time) StateValue {
	return StateValue{string: value, expiresAt: expiresAt}
}

func InitializeState() {
	if initialized {
		log.Fatal("state already initialized")
	}
	// Set initialized to true to prevent reinitialization
	//   we may set here instead of at the end of the function
	//   as we expect initialization failure to be fatal.
	initialized = true
	dir, _ := config.Get("dir")
	dbfilename, _ := config.Get("dbfilename")
	filePath := path.Join(dir, dbfilename)
	f, err := os.Open(path.Join(dir, dbfilename))
	if errors.Is(err, fs.ErrNotExist) {
		log.Printf("RDB file %s does not exist, skipping initialization from RDB file", filePath)
	} else {
		// defer is OK since we don'b care about handling the error here
		defer f.Close()
		if err != nil {
			log.Fatalf("failed to open RDB file: %v", err)
		}
		br := bufio.NewReader(f)
		initializeFromRDB(br)
		return
	}
	log.Println("No RDB file specified or file does not exist, initializing empty state")
	state = map[int64]*stateShard{
		0: {data: make(map[string]StateValue)},
	}
}

func initializeFromRDB(br *bufio.Reader) {
	dbs, err := rdbreader.ReadRDB(br)
	if err != nil {
		log.Fatalf("failed to read RDB: %v", err)
	}
	state = make(map[int64]*stateShard)
	for db, data := range dbs {
		state[db] = &stateShard{data: make(map[string]StateValue)}
		for k, v := range data {
			state[db].data[k] = NewStateValue(v.Value, v.ExpiresAt)
		}
	}
}

type stateShard struct {
	data map[string]StateValue
	mu   sync.RWMutex
}

var state map[int64]*stateShard

// getStateShardForKey for now returns a global state variable.
// This function is intended to allow us to shard the state in the future.
// to achieve greater concurrency.
func getStateShardForDbAndKey(db int64, _ string) *stateShard {
	return state[db]
}

func getAllStateShardsForDb(db int64) []*stateShard {
	return []*stateShard{state[db]}
}

func Set(db int64, key, value string, px int64) {
	ml := getStateShardForDbAndKey(db, key)
	// zero time means no expiry
	var expiresAt time.Time
	if px != -1 {
		expiresAt = time.Now().Add(time.Duration(px) * time.Millisecond)
	}
	ml.mu.Lock()
	ml.data[key] = StateValue{string: value, expiresAt: expiresAt}
	ml.mu.Unlock()
}

func Get(db int64, key string) (string, bool) {
	ml := getStateShardForDbAndKey(db, key)
	ml.mu.RLock()
	v, ok := ml.data[key]
	ml.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !v.expiresAt.IsZero() && v.expiresAt.Before(time.Now()) {
		go tryEvictExpiredKey(db, key)
		return "", false
	}
	return v.string, true
}

func Keys(db int64) []string {
	lengthEstimate := 0
	shards := getAllStateShardsForDb(db)
	for _, shard := range shards {
		lengthEstimate += len(shard.data) + 10
	}
	keys := make([]string, 0, lengthEstimate)
	for _, shard := range shards {
		shard.mu.RLock()
	}
	now := time.Now()
	for _, shard := range shards {
		for k, v := range shard.data {
			if !v.expiresAt.IsZero() && v.expiresAt.Before(now) {
				continue
			}
			keys = append(keys, k)
		}
		shard.mu.RUnlock()
	}
	return keys
}

func tryEvictExpiredKey(db int64, key string) {
	ml := getStateShardForDbAndKey(db, key)
	ml.mu.Lock()
	defer ml.mu.Unlock()
	v, ok := ml.data[key]
	if !ok {
		return
	}
	if !v.expiresAt.IsZero() && v.expiresAt.Before(time.Now()) {
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
	for db := range state {
		for _, ml := range getAllStateShardsForDb(db) {
			if len(ml.data) < evictionSweepMapSizeThreshold {
				continue
			}
			now := time.Now()
			ml.mu.Lock()
			i := 0
			for k, v := range ml.data {
				if !v.expiresAt.IsZero() && v.expiresAt.Before(now) {
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
}

func DummyDumpStateAsString() string {
	bs, _ := hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")
	return string(bs)
}
