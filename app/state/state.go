package state

import (
	"errors"
	"log"
	"sort"
	"strconv"
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

var _ DbValue = (*DbStream)(nil)

func (v *DbStream) Type() string {
	return "stream"
}

var _ sort.Interface = (*DbStream)(nil)

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
