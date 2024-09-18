package replication

import (
	"encoding/json"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type ReplicationInfo struct {
	Role             string `json:"role"`
	MasterReplid     string `json:"master_replid"`
	MasterReplOffset int    `json:"master_repl_offset"`

	masterClient *client.Client `json:"-"`
}

var replicationInfo = ReplicationInfo{
	MasterReplid:     "?",
	MasterReplOffset: -1,
}

func InitializeReplication() {
	replicaof, _ := config.Get("replicaof")
	var err error
	switch replicaof {
	case "":
		err = initializeMaster(&replicationInfo)
	default:
		err = InitializeSlave(&replicationInfo, replicaof)
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
