package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var typeCommand = "TYPE"

func handleType(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	key := sa[1]
	t := state.Type(key)
	return &resp.RESPSimpleString{Value: t}, nil
}
