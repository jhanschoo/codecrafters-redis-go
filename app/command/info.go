package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

var infoCommand = "INFO"

func handleInfo(sa []string, _ int64) (resp.RESP, error) {
	if len(sa) != 2 || sa[1] != "replication" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only INFO replication is supported for the INFO command`}, nil
	}

	replicationInfo, err := replication.GetReplicationInfoAsMap()
	if err != nil {
		return &resp.RESPSimpleError{Value: fmt.Sprintf("Error getting replication info: %v", err)}, nil
	}
	serializedReplicationInfo := utility.SerializeInfo(replicationInfo)

	return &resp.RESPBulkString{Value: serializedReplicationInfo}, nil
}
