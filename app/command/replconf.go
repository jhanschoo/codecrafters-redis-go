package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var replconfCommand = "REPLCONF"

func handleReplconf(sa []string, _ int64) (resp.RESP, error) {
	if len(sa) <= 2 {
		return &resp.RESPSimpleError{Value: `Expected at least 2 arguments for REPLCONF`}, nil
	}
	// dummy implementation
	return resp.RESPSimpleString{Value: "OK"}, nil
}
