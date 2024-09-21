package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var incrCommand = "INCR"

func handleIncr(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	var key = sa[1]
	if ctx.IsReplica && !ctx.IsReplConn {
		return &resp.RESPSimpleError{Value: "READONLY You can't write against a read only replica."}, nil
	}
	var res int64
	if err := state.ExecuteAndReplicateCommand(func() error {
		var err error
		res, err = state.Incr(key)
		return err
		// TODO: change replication command to SET for expiry consistency
	}, ctx.Com); err != nil {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return resp.RESPInteger{Value: res}, nil
}
