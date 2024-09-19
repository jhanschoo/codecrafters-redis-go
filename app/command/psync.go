package command

import (
	"fmt"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var psyncCommand = "PSYNC"

func handlePsync(_ []string, ctx Context) error {
	ri := replication.GetReplicationInfo()

	// write initial response
	res := resp.RESPSimpleString{Value: fmt.Sprintf("FULLRESYNC %s %d", ri.MasterReplid, ri.MasterReplOffset)}
	ctx.Conn.Write([]byte(res.SerializeRESP()))

	// dump state
	// note: in a proper implementation, state.DummyDumpStateAsString() will be replaced with a function that
	// 1. globally read-locks the state
	// 2. dumps the state to an RDB file
	// 3. creates a buffered pipe, and subscribes its producer end listener end to state mutations
	// 4. globally unlocks the state
	// 5. returns a fp to the RDB file and the consumer end of the pipe
	// 6. streams the RDB file to the connection
	// 7. connects the consumer end of the pipe to the connection
	st := state.DummyDumpStateAsString()
	rst := &resp.RESPBulkString{Value: st}
	bs := []byte(rst.SerializeRESP())
	// write without the trailing \r\n
	_, err := ctx.Conn.Write(bs[:len(bs)-2])
	log.Printf("Registering listener for replication")
	replication.RegisterListener(ctx.Conn)

	return err
}
