package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
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
	// dummy implementation
	return resp.RESPInteger{Value: 0}, nil
}
