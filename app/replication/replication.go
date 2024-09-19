package replication

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type ReplicationInfo struct {
	Role             string `json:"role"`
	MasterReplid     string `json:"master_replid"`
	MasterReplOffset int    `json:"master_repl_offset"`

	masterClient *client.Client        `json:"-"`
	listeners    []replicationListener `json:"-"`
}

type replicationListener struct {
	io.Writer
	l *sync.Mutex
}

var replicationInfo = ReplicationInfo{
	MasterReplid:     "?",
	MasterReplOffset: -1,
	listeners:        make([]replicationListener, 0),
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

func RegisterListener(w io.Writer) {
	replicationInfo.listeners = append(replicationInfo.listeners, replicationListener{w, &sync.Mutex{}})
}

func WriteToAllListeners(p []byte) {
	for _, w := range replicationInfo.listeners {
		writeToListener(p, w)
	}
}

func writeToListener(p []byte, w replicationListener) {
	w.l.Lock()
	w.Write(p)
	w.l.Unlock()
}
