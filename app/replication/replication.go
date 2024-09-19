package replication

import (
	"encoding/json"
	"io"
	"log"
	"sync"
	"sync/atomic"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type ReplicationInfo struct {
	Role             string         `json:"role"`
	MasterReplid     string         `json:"master_replid"`
	MasterReplOffset int            `json:"master_repl_offset"`
	MasterClient     *client.Client `json:"-"`

	listeners   []io.Writer `json:"-"`
	listenersMu *sync.Mutex `json:"-"`
}

var replicationInfo = ReplicationInfo{
	MasterReplid:     "?",
	MasterReplOffset: -1,
	listeners:        make([]io.Writer, 0),
	listenersMu:      &sync.Mutex{},
}

func InitializeReplication() {
	replicaof, _ := config.Get("replicaof")
	var err error
	switch replicaof {
	case "":
		err = initializeMaster()
	default:
		err = initializeSlave(replicaof)
	}
	if err != nil {
		log.Fatalf("Failed to initialize replication: %v", err)
	}
}

func GetReplicationInfo() ReplicationInfo {
	return replicationInfo
}

func GetReplicationInfoAsJson() ([]byte, error) {
	return json.Marshal(replicationInfo)
}

func GetReplicationInfoAsMap() (map[string]interface{}, error) {
	jsonBytes, err := GetReplicationInfoAsJson()
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Contract: once a writer is registered, no other part of the code should write to it except indirectly through this package.
func RegisterListener(w io.Writer) {
	replicationInfo.listeners = append(replicationInfo.listeners, w)
}

func GetListenersCount() int {
	replicationInfo.listenersMu.Lock()
	defer replicationInfo.listenersMu.Unlock()
	return len(replicationInfo.listeners)
}

func ExecuteAndWriteToListenersAtomically(f func() error, bs []byte) {
	replicationInfo.listenersMu.Lock()
	l := len(replicationInfo.listeners)
	if err := f(); err != nil || l == 0 {
		replicationInfo.listenersMu.Unlock()
		return
	}
	var counter atomic.Int64
	counter.Store(int64(l))
	listeners := make([]io.Writer, l)
	copy(listeners, replicationInfo.listeners)
	for _, w := range replicationInfo.listeners {
		go writeToListener(bs, w, &counter)
	}
}

func writeToListener(p []byte, w io.Writer, counter *atomic.Int64) {
	w.Write(p)
	i := counter.Add(-1)
	if i == 0 {
		replicationInfo.listenersMu.Unlock()
	}
}
