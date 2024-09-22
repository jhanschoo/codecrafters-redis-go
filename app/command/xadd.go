package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var xaddCommand = "XADD"

func handleXadd(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) <= 4 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected an at least 4-element array"}, nil
	}
	key, id := sa[1], sa[2]
	if ctx.IsReplica && !ctx.IsReplConn {
		return &resp.RESPSimpleError{Value: "READONLY You can't write against a read only replica."}, nil
	}
	coms := []resp.RESP{ctx.Com}
	if err := state.ExecuteAndReplicateCommand(func() ([]resp.RESP, error) {
		var err error
		id, err = state.Xadd(key, id, sa[3:])
		return coms, err
	}); err != nil {
		if err == state.ErrorNone {
			return resp.NullLit, nil
		}
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return &resp.RESPBulkString{Value: id}, nil
}
