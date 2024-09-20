package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var infoCommand = "INFO"

func handleInfo(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) != 2 || sa[1] != "replication" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only INFO replication is supported for the INFO command`}, nil
	}

	replInfo := state.GetReplInfo().Serialize()

	return &resp.RESPBulkString{Value: replInfo}, nil
}
