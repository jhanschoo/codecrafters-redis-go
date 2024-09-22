package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

var discardCommand = "DISCARD"

var ErrorDiscardNotTransaction = &resp.RESPSimpleError{Value: "ERR DISCARD without MULTI"}

func handleDiscard(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 1 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 1-element array"}, nil
	}
	if !ctx.Queued.IsActive() {
		return ErrorDiscardNotTransaction, nil
	}
	_ = ctx.Queued.RetrieveComs()
	return resp.OkLit, nil
}
