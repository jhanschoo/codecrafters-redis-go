package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var multiCommand = "MULTI"

func handleMulti(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 1 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 1-element array"}, nil
	}
	ok := ctx.Queued.Initialize()
	if !ok {
		return &resp.RESPSimpleError{Value: "MULTI calls can not be nested"}, nil
	}
	return resp.OkLit, nil
}
