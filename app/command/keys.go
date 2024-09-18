package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func handleKeys(db int64, sa []string) resp.RESP {
	if len(sa) != 2 || sa[1] != "*" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only KEYS "*" is supported for the KEYS command`}
	}
	keys := state.Keys(db)
	return resp.ParseStringSlice(keys)
}
