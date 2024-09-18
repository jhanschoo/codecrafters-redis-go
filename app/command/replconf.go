package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleReplconf(_ int64, sa []string) resp.RESP {
	if len(sa) <= 2 {
		return &resp.RESPSimpleError{Value: `Expected at least 2 arguments for REPLCONF`}
	}
	switch sa[1] {
	case "listening-port":
		return resp.RESPSimpleString{Value: "OK"}
	case "capa":
		return resp.RESPSimpleString{Value: "OK"}
	default:
		return &resp.RESPSimpleError{Value: `Unsupported REPLCONF command`}
	}
}
