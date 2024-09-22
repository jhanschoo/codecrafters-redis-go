package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var execCommand = "EXEC"

var ErrorExecNotTransaction = &resp.RESPSimpleError{Value: "ERR EXEC without MULTI"}

func handleExec(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 1 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	if !ctx.Queued.IsActive() {
		return ErrorExecNotTransaction, nil
	}
	_ = ctx.Queued.RetrieveComs()
	res := &resp.RESPArray{Value: make([]resp.RESP, 0)}
	return res, nil
}
