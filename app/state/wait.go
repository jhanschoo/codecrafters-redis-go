package state

import (
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/utility"
)

type waitState struct {
	cond            *sync.Cond
	numReplicas     int64
	numAcked        int64
	ackThreshold    int64
	offsetThreshold int64
	done            bool
	l               *sync.Mutex
}

func newWaitState(minRepl, offsetThreshold int64) *waitState {
	l := &sync.Mutex{}
	return &waitState{
		cond:            sync.NewCond(l),
		numReplicas:     -1,
		numAcked:        0,
		ackThreshold:    minRepl,
		offsetThreshold: offsetThreshold,
		l:               l,
	}
}

func HandleWait(minRepl int64, timeout time.Duration) int64 {
	if minRepl <= 0 {
		return 0
	}
	// SPECIAL CASE for codecrafters: if the replication stream has offset 0, we can return immediately
	if state.MasterReplOffset.Load() == 0 {
		replicasMu.Lock()
		l := len(replicas)
		replicasMu.Unlock()
		return int64(l)
	}
	ws := newWaitState(minRepl, state.MasterReplOffset.Load())
	go utility.Timeout(timeout, ws.l, ws.cond, func() bool {
		ws.done = true
		return true
	})

	propagateGetAck(ws)
	// at this point, numReplicas is set on ws

	for {
		if acked := handleWaitAux(ws); acked != -1 {
			return acked
		}
	}
}

func handleWaitAux(ws *waitState) int64 {
	ws.l.Lock()
	ws.cond.Wait()
	defer ws.l.Unlock()
	if ws.numAcked >= ws.ackThreshold || ws.done {
		ws.done = true
		// ws.numAcked is evaluated before ws.l.Unlock(), so no explicit intermediate variable is needed
		return ws.numAcked
	}
	return -1
}
