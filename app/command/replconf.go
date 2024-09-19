package command

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var replconfCommand = "REPLCONF"

func handleReplconf(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) <= 2 {
		return &resp.RESPSimpleError{Value: `Expected at least 2 arguments for REPLCONF`}, nil
	}
	// dummy implementation
	switch sa[1] {
	case "listening-port":
		return resp.RESPSimpleString{Value: "OK"}, nil
	case "capa":
		return resp.RESPSimpleString{Value: "OK"}, nil
	case "GETACK":
		ret := []string{"REPLCONF", "ACK", strconv.FormatInt(ctx.BytesProcessed, 10)}
		return resp.EncodeStringSlice(ret), nil
	default:
		return &resp.RESPSimpleError{Value: `Unsupported REPLCONF command`}, nil
	}
}
