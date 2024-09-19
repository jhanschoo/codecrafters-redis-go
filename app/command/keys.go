package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var keysCommand = "KEYS"

func handleKeys(sa []string, db int64) (resp.RESP, error) {
	if len(sa) != 2 || sa[1] != "*" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only KEYS "*" is supported for the KEYS command`}, nil
	}
	keys := state.Keys(db)
	return resp.EncodeStringSlice(keys), nil
}
