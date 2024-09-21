package state

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

// dummy implementation of PSYNC; will always return a FULLRESYNC command, and assume no history
func HandlePsync(r *respreader.BufferedRESPConnReader) error {
	log.Println("state.HandlePsync: started")
	// we set up readers and writers in the following manner:
	// 1. replica.w is bufio.Writer, so that `forwardCommands` can write to it without blocking. Moreover, it does not wrap around another buffer
	// 2. underlying replica.w is a pipe writer pw, piped to pipe reader pr
	// 3. pr is the second of the two readers in the io.MultiReader; the first reader is the dump of the state
	// 4. io.Copy is used to copy from the multi-reader to the replica as indefinitely; first the dump, then the propagations are copied
	// Recall that io.Copy works as follows: it reads into a buffer in a blocking manner, and writes to the writer in a blocking manner, and repeats until an error is encountered
	// 5. when `forwardCommands` encounters a sync command, it writes GETACK to replica.w and then flushes it. This blocks it until io.Copy eventually reads it. `forwardCommands` then proceeds to block itself on reading from the connection, and it will have data to read once the replica's response reaches connection
	pr, pw := io.Pipe()
	bpr := bufio.NewReader(pr)
	replica := newReplica(r, pw)

	state.PropagateMu.Lock()
	replid := state.MasterReplid
	replOffset := state.MasterReplOffset.Load()
	state.DbMu.RLock()
	dumpr := unsafeInitiateDump()
	state.DbMu.RUnlock()
	replica.registerSelf()
	go replica.forwardCommands()
	// at this point, the replica is registered, with a goroutine forwarding propagations into the write buffer, so we can release the lock
	state.PropagateMu.Unlock()
	mr := io.MultiReader(dumpr, bpr)

	// write initial response
	res := resp.RESPSimpleString{Value: fmt.Sprintf("FULLRESYNC %s %d", replid, replOffset)}
	r.Write([]byte(res.SerializeRESP()))
	// tee := io.MultiWriter(r, log.Writer())

	if _, err := io.Copy(r, mr); err != nil {
		log.Printf("Error piping to replica: %v", err)
	}
	replica.unregisterSelf()
	return nil
}
