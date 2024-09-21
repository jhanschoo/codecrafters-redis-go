package command

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var setCommand = "SET"

func handleSet(sa []string, ctx Context) (resp.RESP, error) {
	var (
		key   string
		value string
		px    int64
		err   error
	)
	switch len(sa) {
	case 3:
		key = sa[1]
		value = sa[2]
		px = -1
	case 5:
		if strings.ToUpper(sa[3]) != "PX" {
			return &resp.RESPSimpleError{Value: "Invalid input: expected PX as 4th element"}, nil
		}
		key = sa[1]
		value = sa[2]
		px, err = strconv.ParseInt(sa[4], 10, 64)
		if err != nil {
			return &resp.RESPSimpleError{Value: "Invalid input: expected integer as 5th element"}, nil
		}
	default:
		return &resp.RESPSimpleError{Value: "Invalid input: expected 3 or 5-element array"}, nil
	}
	if ctx.IsReplica && !ctx.IsReplConn {
		return &resp.RESPSimpleError{Value: "READONLY You can't write against a read only replica."}, nil
	}
	coms := []resp.RESP{ctx.Com}
	if err := state.ExecuteAndReplicateCommand(func() ([]resp.RESP, error) {
		return coms, state.Set(key, value, px)
	}); err != nil {
		return nil, err
	}
	return respOk, nil
}
