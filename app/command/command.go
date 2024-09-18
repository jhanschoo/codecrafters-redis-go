package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var commandHandlers = map[string]func(db int64, sa []string) resp.RESP{
	"PING":   handlePing,
	"ECHO":   handleEcho,
	"SET":    handleSet,
	"GET":    handleGet,
	"CONFIG": handleConfigCommands,
	"KEYS":   handleKeys,
	"INFO":   handleInfo,
}

func Handle(db int64, sa []string) resp.RESP {
	if len(sa) == 0 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected non-empty array of bulk strings"}
	}
	handler, ok := commandHandlers[sa[0]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Unsupported command " + sa[0]}
	}
	return handler(db, sa)
}
