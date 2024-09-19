package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var getCommand = "GET"

func handleGet(sa []string, db int64) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	key := sa[1]
	value, ok := state.Get(db, key)
	if !ok {
		return &resp.RESPNull{CompatibilityFlag: 1}, nil
	}
	return &resp.RESPBulkString{Value: value}, nil
}
