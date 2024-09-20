package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var configCommand = "CONFIG"

var configCommandHandlers = map[string]standardSubhandler{
	"GET": handleConfigGet,
}

func handleConfigCommands(sa []string, ctx Context) (resp.RESP, error) {
	if len(sa) <= 1 {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG command"}, nil
	}
	handler, ok := configCommandHandlers[strings.ToUpper(sa[1])]
	if !ok {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG " + sa[1] + " command"}, nil
	}
	return handler(sa, ctx)

}
