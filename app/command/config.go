package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var configCommandHandlers = map[string]func(sa []string) resp.RESP{
	"GET": handleConfigGet,
}

func handleConfigCommands(sa []string) resp.RESP {
	if len(sa) <= 1 {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG command"}
	}
	handler, ok := configCommandHandlers[sa[1]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG " + sa[1] + " command"}
	}
	return handler(sa)

}
