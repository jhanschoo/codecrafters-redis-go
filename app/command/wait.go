package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var waitCommand = "WAIT"

func handleWait(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 3 {
		return &resp.RESPSimpleError{Value: `Expected 3 arguments for WAIT`}, nil
	}
	// dummy implementation
	return resp.RESPInteger{Value: int64(replication.GetListenersCount())}, nil
}
