package command

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var psyncCommand = "PSYNC"

var (
	ErrorIsReplica           = errors.New("received PSYNC command on replica node")
	ErrorNotExpectedToReturn = &resp.RESPSimpleError{Value: "PSYNC command is expected to live indefinitely long"}
)

func handlePsync(_ []string, ctx Context) (resp.RESP, error) {
	if ctx.IsReplica {
		return nil, ErrorIsReplica
	}
	// The following function call is expected to live indefinitely long
	return ErrorNotExpectedToReturn, state.HandlePsync(ctx.Reader)
}
