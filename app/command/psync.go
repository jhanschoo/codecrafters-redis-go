package command

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func handlePsync(_ int64, _ []string) (resp.RESP, connHandler) {
	ri := replication.GetReplicationInfo()
	next := func(c net.Conn) {
		st := state.DummyDumpStateAsString()
		rst := &resp.RESPBulkString{Value: st}
		bs := []byte(rst.SerializeRESP())
		// Remove the trailing \r\n
		c.Write(bs[:len(bs)-2])
	}

	return &resp.RESPSimpleString{Value: fmt.Sprintf("FULLRESYNC %s %d", ri.MasterReplid, ri.MasterReplOffset)}, &next
}
