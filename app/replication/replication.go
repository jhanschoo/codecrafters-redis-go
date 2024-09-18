package replication

import (
	"encoding/json"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

type ReplicationInfo struct {
	Role             string `json:"role"`
	Replicaof        string `json:"-"`
	MasterReplid     string `json:"master_replid"`
	MasterReplOffset int    `json:"master_repl_offset"`
}

var replicationInfo ReplicationInfo

func InitializeReplication() {
	replicaof, _ := config.Get("replicaof")
	switch replicaof {
	case "":
		replicationInfo.Role = "master"
		replicationInfo.MasterReplid = utility.RandomAlphaNumericString(40)
	default:
		replicationInfo.Role = "slave"
		replicationInfo.Replicaof = replicaof
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
