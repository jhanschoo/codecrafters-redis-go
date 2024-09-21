package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var getCommand = "GET"

func handleGet(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	key := sa[1]
	value, err := state.Get(key)
	if err == state.ErrorNone {
		return respNull, nil
	}
	if err == state.ErrorWrongType {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	if err != nil {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return &resp.RESPBulkString{Value: value}, nil
}
