package command

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var waitCommand = "WAIT"

// To implement this next stage, we must maintain clocks that count the bytes published by the server to replicas since initialization.
// When a replica connects with PSYNC, at the atomic point where the master snapshots the RDB dump, the master records the clock value and associates it with the replica that has just connected, in the listener entry object. In addition, it creates a channel for asynchronous yet ordered communication with the replica, and a lock that needs to be acquired before writing to the channels.
// From hereon, the threads managing the replicas will solely behave by reading from the channel and writing to the replica.
// Whenever the master is compelled to communicate with its replicas via the listener, the thread initiating the communication will acquire the lock, write to the channel, and release the lock.
// For the WAIT command, the command will
// 1. spawn a sleeper thread that sleeps for the specified timeout, then returns
// 2. spawn a verification thread that returns when enough replicas are known to be sufficiently synced, or if the main thread tells it to give up
// The main thread then selects conditional on either thread responding.
// When woken, the main thread signals the verification thread to give up, then returns.
// For the verification thread, it acquires the lock, notes down the master's clock, then sends the REPLCONF GETACK * to every replica channel. th spawn a worker thread. the master will acquire the lock, write to the channel. The thread managing the replica must recognize the this is a message that expects a response, and writes back to the channel once the replica has responded.
func handleWait(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 3 {
		return &resp.RESPSimpleError{Value: `Expected 3 arguments for WAIT`}, nil
	}
	if ctx.IsReplica {
		return &resp.RESPSimpleError{Value: `READONLY You can't WAIT while connected to a read-only replica.`}, nil
	}
	minRepl, err := strconv.ParseInt(sa[1], 10, 64)
	if err != nil {
		return &resp.RESPSimpleError{Value: `Invalid input: expected integer as 2nd element`}, nil
	}
	timeoutInMs, err := strconv.ParseInt(sa[2], 10, 64)
	if err != nil {
		return &resp.RESPSimpleError{Value: `Invalid input: expected integer as 3rd element`}, nil
	}
	timeout := time.Duration(timeoutInMs) * time.Millisecond
	acks := state.HandleWait(minRepl, timeout)
	// dummy implementation
	return resp.RESPInteger{Value: acks}, nil
}
