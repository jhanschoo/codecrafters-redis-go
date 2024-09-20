package command

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var psyncCommand = "PSYNC"

var ErrorIsReplica = errors.New("Received PSYNC command on replica node")

func handlePsync(_ []string, ctx Context) error {
	if ctx.IsReplica {
		return ErrorIsReplica
	}
	// The following function call is expected to live indefinitely long
	return state.HandlePsync(ctx.Reader)
}
