package state

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

var initialized = false

// state definition and getters

type State struct {
	// Replication info
	// Note: locking, etc. would be necessary if we expect failover scenarios
	Role         string `json:"role"`
	MasterReplid string `json:"master_replid"`

	// Note after initialization the different sources of mutation for MasterReplOffset depending on the role
	// For master, it is mutated by several sources including mutations to the database and liveness checks
	// For replica, it is mutated by the master connection only
	MasterReplOffset atomic.Int64   `json:"master_repl_offset"`
	MasterClient     *client.Client `json:"-"`

	// Database state
	Db Db `json:"-"`
	// While a thread holds the lock, no other thread is expected to mutate the database; and the thread itself may not mutate the database
	// if it acquired a read lock
	DbMu sync.RWMutex `json:"-"`

	// Replication state
	// While a thread holds the lock, no other thread is expected to mutate the database in a way that would require propagation, or to mutate the replication stream, or to add replicas (replicas may be removed, and mutations to the database may be made if they do not require propagation)
	// For deadlock avoidance, the lock should be acquired before the database lock, and the lock should be acquired before the replica lock
	PropagateMu sync.Mutex `json:"-"`
}

var state = State{}

func IsReplica() bool {
	return state.Role == "slave"
}

func IsReplConn(r *respreader.BufferedRESPConnReader) bool {
	return state.MasterClient != nil && r == state.MasterClient.BufferedRESPConnReader
}

func MasterReplid() string {
	return state.MasterReplid
}

func ReplOffset() int64 {
	return state.MasterReplOffset.Load()
}

func IncrOffset(by int64) {
	state.MasterReplOffset.Add(by)
}

func CasOffset(old, new int64) bool {
	return state.MasterReplOffset.CompareAndSwap(old, new)
}

func GetReplInfo() utility.Info {
	return utility.Info{
		"role":               utility.InfoString(state.Role),
		"master_replid":      utility.InfoString(state.MasterReplid),
		"master_repl_offset": utility.InfoString(strconv.FormatInt(state.MasterReplOffset.Load(), 10)),
	}
}

func MasterClient() *client.Client {
	return state.MasterClient
}

// db struct definitions
var (
	ErrorNone      = errors.New("ERR no such key")
	ErrorWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
)

type Db = map[string]DbValue

type DbValue interface {
	Type() string
}

type DefinitelyExpirer interface {
	IsDefinitelyExpiredAt(t time.Time) bool
}

type DbNone struct{}

var _ DbValue = (*DbNone)(nil)

func (v *DbNone) Type() string {
	return "none"
}

var NoneValue = DbNone{}

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

type DbStream struct {
	data []DbStreamEntry
}

type DbStreamEntry struct {
	ms     int64
	seq    int64
	fields []string
}

func NewDbStreamEntry(id string, fields []string) (DbStreamEntry, error) {
	ms, seq, err := parseStreamEntryId(id)
	if err != nil {
		return DbStreamEntry{}, err
	}
	return DbStreamEntry{ms, seq, fields}, nil
}

func parseStreamEntryId(id string) (ms, seq int64, err error) {
	parts := strings.Split(id, "-")
	ms, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	seq, err = strconv.ParseInt(parts[1], 10, 64)
	return
}

func (e DbStreamEntry) Id() string {
	return strconv.FormatInt(e.ms, 10) + "-" + strconv.FormatInt(e.seq, 10)
}

var _ DbValue = (*DbStream)(nil)

func (v *DbStream) Type() string {
	return "stream"
}

var _ sort.Interface = (*DbStream)(nil)

func (v *DbStream) Len() int {
	return len(v.data)
}

func (v *DbStream) Less(i, j int) bool {
	return v.data[i].ms < v.data[j].ms || (v.data[i].ms == v.data[j].ms && v.data[i].seq < v.data[j].seq)
}

func (v *DbStream) Swap(i, j int) {
	v.data[i], v.data[j] = v.data[j], v.data[i]
}

func (v *DbStream) ValidNextId(id string) error {
	ms, seq, err := parseStreamEntryId(id)
	if err != nil {
		return ErrorInvalidIdFormat
	}
	if !(ms >= 0 && seq > 0) {
		return ErrorInvalidIdValue
	}
	if len(v.data) == 0 {
		return nil
	}
	if last := v.data[len(v.data)-1]; last.ms > ms || (last.ms == ms && last.seq >= seq) {
		return ErrorInvalidNewId
	}
	return nil
}

// high-level state management

// Initialization

func Initialize() {
	if initialized {
		log.Fatalf("state already initialized")
	}
	initialized = true
	replicaof := config.Get("replicaof")
	if replicaof != "" {
		initializeReplica()
	} else {
		initializeMaster()
	}
}

// Db operations

// Initialization and replication operations

func UnsafeSet(key, value string, expiresAt time.Time) {
	state.Db[key] = &DbString{string: value, expiresAt: expiresAt}
}

func UnsafeResetDbWithSizeHint(sizeHint int64) {
	state.Db = make(map[string]DbValue, sizeHint)
}

// General operations

func Type(key string) string {
	state.DbMu.RLock()
	v, ok := state.Db[key]
	state.DbMu.RUnlock()
	if !ok {
		return NoneValue.Type()
	}
	return v.Type()
}

func Keys() []string {
	keys := make([]string, 0, len(state.Db)+len(state.Db)/10)
	now := time.Now()
	state.DbMu.RLock()
	defer state.DbMu.RUnlock()
	for k, v := range state.Db {
		if w, ok := v.(DefinitelyExpirer); ok && w.IsDefinitelyExpiredAt(now) {
			go TryEvictExpiredKey(k)
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

// String operations

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

// Stream operations
var (
	ErrorInvalidIdFormat = errors.New("ERR Invalid stream ID format")
	ErrorInvalidIdValue  = errors.New("ERR The ID specified in XADD must be greater than 0-0")
	ErrorInvalidNewId    = errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
)

func Xadd(key string, id string, fields []string) (string, error) {
	state.DbMu.Lock()
	defer state.DbMu.Unlock()
	v, ok := state.Db[key]
	if !ok {
		v = &DbStream{data: make([]DbStreamEntry, 0, 1)}
		state.Db[key] = v
	}
	stream, ok := v.(*DbStream)
	if !ok {
		return "", ErrorWrongType
	}
	if err := stream.ValidNextId(id); err != nil {
		return "", err
	}
	e, err := NewDbStreamEntry(id, fields)
	if err != nil {
		return "", err
	}
	stream.data = append(stream.data, e)
	return id, nil
}

// replication operations

// Note that replicas are expected to execute this function just as with the master, except that they have no replicas of their own to propagate to
func ExecuteAndReplicateCommand(f func() error, cmd resp.RESP) error {
	cmdstr := cmd.SerializeRESP()
	delta := int64(len(cmdstr))
	state.PropagateMu.Lock()
	defer state.PropagateMu.Unlock()
	if err := f(); err != nil {
		return err
	}
	IncrOffset(delta)
	unsafePropagate(replMessage{s: cmdstr, ack: nil})
	return nil
}
