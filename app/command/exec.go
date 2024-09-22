package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var execCommand = "EXEC"

var ErrorExecNotTransaction = &resp.RESPSimpleError{Value: "ERR EXEC without MULTI"}

func handleExec(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) != 1 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 1-element array"}, nil
	}
	if !ctx.Queued.IsActive() {
		return ErrorExecNotTransaction, nil
	}
	coms := ctx.Queued.RetrieveComs()
	res := &resp.RESPArray{Value: make([]resp.RESP, len(coms))}
	for i, com := range coms {
		r, err := ctx.Handle(Context{
			Reader:        ctx.Reader,
			IsReplica:     ctx.IsReplica,
			IsReplConn:    ctx.IsReplConn,
			Com:           com,
			Queued:        ctx.Queued,
			InTransaction: true,
		})
		if err != nil {
			return nil, err
		}
		res.Value[i] = r
	}
	return res, nil
}
