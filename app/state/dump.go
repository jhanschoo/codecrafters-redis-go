package state

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var dummyDumpStateBytes, _ = hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")

// unsafeInitiateDump assumes that the caller has acquired locks to the state.
// It
// 1. sets up the state for a dump operation
// 2. spawns a goroutine to dump the state at the time it was locked to an RDB file (alternatively, the dump can be managed by communicating with a daemon that otherwise periodically dumps the state)
// 3. returns a placeholder io.Reader that will be used to stream the RDB file to the replica
// 4. when the dump is complete,
// 5. it spawns yet another goroutine that manage streaming the dump to the returned io.Reader, and when done performs further management of the dump if necessary
// 6. the dump operation goroutine is expected to renormalize the state after the dump is complete
// However, we currently provide a dummy implementation that assumes an empty state with no history.
func unsafeInitiateDump() io.Reader {
	rst := &resp.RESPBulkString{Value: string(dummyDumpStateBytes)}
	bs := []byte(rst.SerializeRESP())
	bs = bs[0 : len(bs)-2] // remove the trailing \r\n
	return bytes.NewReader(bs)
}
