package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var commandHandlers = map[string]func(sa []string) resp.RESP{
	"PING": handlePing,
	"ECHO": handleEcho,
	"SET":  handleSet,
	"GET":  handleGet,
}

func Handle(sa []string) resp.RESP {
	if len(sa) == 0 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected non-empty array of bulk strings"}
	}
	handler, ok := commandHandlers[sa[0]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Invalid command"}
	}
	return handler(sa)
}
