package state

import (
	"bufio"
	"log"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

type reader = respreader.BufferedRESPConnReader

type replMessage struct {
	s     string
	isAck bool
}

type replica struct {
	r  *reader
	w  *bufio.Writer
	dc chan replMessage
	cc chan byte
}

var replicas = make(map[*replica]bool)

// while a goroutine holds the lock, no other goroutine is expected to change the set of replicas
// note, of course, that goroutines may be interacting with individual values in the map concurrently
var replicasMu = sync.Mutex{}

func newReplica(r *reader, w *bufio.Writer) *replica {
	rep := &replica{
		r:  r,
		w:  w,
		dc: make(chan replMessage, 1),
		cc: make(chan byte),
	}
	return rep
}

func (r *replica) registerSelf() {
	replicasMu.Lock()
	replicas[r] = true
	replicasMu.Unlock()
}

// This is used by
// 1. ExecuteAndReplicateCommand to propagate mutations to replicas
// 2. Wait (TODO: better name) to propagate GETACKs to replicas
func unsafePropagate(msg replMessage) {
	// channels are expected to have single buffer and not block on writes
	replicasMu.Lock()
	defer replicasMu.Unlock()
	for r := range replicas {
		r.dc <- msg
	}
}

// forwardCommands runs synchronously on the thread that handled the PSYNC command, and the calling function is expected to do cleanup
func (r *replica) forwardCommands() {
	for {
		select {
		case msg := <-r.dc:
			if msg.isAck {
				// todo: handle ack
				continue
			}
			n, err := r.w.Write([]byte(msg.s))
			if err != nil {
				log.Println("Error writing to replica:", err)
				r.unregisterSelf()
			}
			if n != len(msg.s) {
				log.Println("Error writing to replica: short write")
				r.unregisterSelf()
			}
			// for codecrafters expectations, we flush after every write
			if err := r.w.Flush(); err != nil {
				log.Println("Error flushing to replica:", err)
				r.unregisterSelf()
			}
		// not expected to happen
		case <-r.cc:
			return
		}
	}
}

func (r *replica) unregisterSelf() {
	replicasMu.Lock()
	delete(replicas, r)
	replicasMu.Unlock()
}
