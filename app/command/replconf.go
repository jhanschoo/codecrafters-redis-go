package command

import (
	"log"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var replconfCommand = "REPLCONF"

func handleReplconf(sa []string, ctx Context) (resp.RESP, error) {
	var ret resp.RESP
	if len(sa) <= 2 {
		ret = &resp.RESPSimpleError{Value: `Expected at least 2 arguments for REPLCONF`}
		return ret, nil
	}
	// dummy implementation
	switch sa[1] {
	case "listening-port":
		return resp.OkLit, nil
	case "capa":
		return resp.OkLit, nil
	case "GETACK":
		log.Println("ReplOffset at GETACK handling(...):", ctx.ReplOffset)
		ret := []string{"REPLCONF", "ACK", strconv.FormatInt(ctx.ReplOffset, 10)}
		return resp.EncodeStringSlice(ret), nil
	default:
		return &resp.RESPSimpleError{Value: `Unsupported REPLCONF command`}, nil
	}
}
