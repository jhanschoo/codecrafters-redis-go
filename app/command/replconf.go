package command

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var replconfCommand = "REPLCONF"

func handleReplconf(sa []string, ctx Context) error {
	var ret resp.RESP
	if len(sa) <= 2 {
		ret = &resp.RESPSimpleError{Value: `Expected at least 2 arguments for REPLCONF`}
		writeRESP(ctx.Reader.Conn, ret)
		return nil
	}
	// dummy implementation
	switch sa[1] {
	case "listening-port":
		writeRESP(ctx.Reader.Conn, resp.OkLit)
		return nil
	case "capa":
		writeRESP(ctx.Reader.Conn, resp.OkLit)
		return nil
	case "GETACK":
		ret := []string{"REPLCONF", "ACK", strconv.FormatInt(ctx.ReplOffset, 10)}
		writeRESP(ctx.Reader.Conn, resp.EncodeStringSlice(ret))
		return nil
	default:
		writeRESP(ctx.Reader.Conn, &resp.RESPSimpleError{Value: `Unsupported REPLCONF command`})
		return nil
	}
}
