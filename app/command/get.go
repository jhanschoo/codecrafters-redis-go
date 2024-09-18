package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func handleGet(db int64, sa []string) resp.RESP {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}
	}
	key := sa[1]
	value, ok := state.Get(db, key)
	if !ok {
		return &resp.RESPNull{CompatibilityFlag: 1}
	}
	return &resp.RESPBulkString{Value: value}
}
