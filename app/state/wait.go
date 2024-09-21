package state

import (
	"sync"
	"time"
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
	l := &sync.Mutex{}
	ws := &waitState{
		cond:            sync.NewCond(l),
		numReplicas:     -1,
		numAcked:        0,
		ackThreshold:    minRepl,
		offsetThreshold: state.MasterReplOffset.Load(),
		l:               l,
	}

	go handleWaitTimeout(timeout, ws)
	broadcastGetAck(ws)
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

func handleWaitTimeout(timeout time.Duration, ws *waitState) {
	// if timeout is 0 the command waits indefinitely and we can return immediately
	if timeout == 0 {
		return
	}
	time.Sleep(timeout)
	ws.l.Lock()
	ws.done = true
	ws.l.Unlock()
	ws.cond.Broadcast()
}
