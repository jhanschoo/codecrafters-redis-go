package state

import (
	"io"
	"log"
	"strconv"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

type reader = respreader.BufferedRESPConnReader

type replMessage struct {
	s   string
	ack func(resp.RESP) bool
}

type replica struct {
	r  *reader
	w  io.Writer
	dc chan replMessage
}

var replicas = make(map[*replica]bool)

// while a goroutine holds the lock, no other goroutine is expected to change the set of replicas
// note, of course, that goroutines may be interacting with individual values in the map concurrently
var replicasMu = sync.Mutex{}

func newReplica(r *reader, w io.Writer) *replica {
	rep := &replica{
		r:  r,
		w:  w,
		dc: make(chan replMessage),
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
	log.Printf("unsafePropagate: propagating %s to %d replicas\n", strconv.Quote(msg.s), len(replicas))
	for r := range replicas {
		r.dc <- msg
	}
}

var getAckString = resp.EncodeStringSlice([]string{"REPLCONF", "GETACK", "*"}).SerializeRESP()
var getAckStringLen = int64(len(getAckString))

func propagateGetAck(ws *waitState) {
	LockPropagateMu()
	defer UnlockPropagateMu()
	ws.l.Lock()
	ws.numReplicas = int64(len(replicas))
	ws.l.Unlock()
	// note that we perform the sync even if there are not enough replicas to meet the ack threshold
	IncrOffset(getAckStringLen)
	unsafePropagate(replMessage{s: getAckString, ack: func(res resp.RESP) bool {
		sa, ok := resp.DecodeStringSlice(res)
		if !ok || len(sa) != 3 || sa[0] != "REPLCONF" || sa[1] != "ACK" {
			return false
		}
		replicaOffset, err := strconv.ParseInt(sa[2], 10, 64)
		if err != nil {
			return false
		}
		if replicaOffset >= ws.offsetThreshold {
			ws.l.Lock()
			ws.numAcked++
			if ws.numAcked >= ws.ackThreshold {
				ws.cond.Broadcast()
			}
			ws.l.Unlock()
			return true
		}
		return false
	}})
}

// forwardCommands is a long-lived function that should run synchronously on the thread that handled the PSYNC command. it
// 1. spawns a goroutine to synchronously handle reads
// 2. then devotes itself to synchronously handling writes (acks)
func (r *replica) forwardCommands() {
	rhs := make([]func(resp.RESP) bool, 0)
	rhsMu := sync.Mutex{}
	go func() {
		for {
			res, err := r.r.ReadRESP()
			log.Printf("forwardCommands %v: Received ack from replica: %s", r, strconv.Quote(res.SerializeRESP()))
			if err != nil {
				log.Println("Error reading from replica:", err)
				r.unregisterSelf()
			}
			rhsMu.Lock()
			for i := 0; i < len(rhs); i++ {
				if rhs[i](res) {
					rhs = append(rhs[:i], rhs[i+1:]...)
					i--
				}
			}
			rhsMu.Unlock()
		}
	}()
	for {
		msg := <-r.dc
		log.Printf("forwardCommands %v: forwarding %s\n", r, strconv.Quote(msg.s))
		if msg.ack != nil {
			rhsMu.Lock()
			rhs = append(rhs, msg.ack)
			rhsMu.Unlock()
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
	}
}

func (r *replica) unregisterSelf() {
	replicasMu.Lock()
	delete(replicas, r)
	replicasMu.Unlock()
}
