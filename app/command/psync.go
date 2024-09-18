package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handlePsync(_ int64, _ []string) resp.RESP {
	ri := replication.GetReplicationInfo()

	return &resp.RESPSimpleString{Value: fmt.Sprintf("FULLRESYNC %s %d", ri.MasterReplid, ri.MasterReplOffset)}
}
