package command

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type connHandler = *func(net.Conn)

var commandHandlers = map[string]func(db int64, sa []string) (resp.RESP, connHandler){
	"PING":     basic(handlePing),
	"ECHO":     basic(handleEcho),
	"SET":      basic(handleSet),
	"GET":      basic(handleGet),
	"CONFIG":   basic(handleConfigCommands),
	"KEYS":     basic(handleKeys),
	"INFO":     basic(handleInfo),
	"REPLCONF": basic(handleReplconf),
	"PSYNC":    handlePsync,
}

func Handle(db int64, sa []string) (resp.RESP, connHandler) {
	if len(sa) == 0 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected non-empty array of bulk strings"}, nil
	}
	handler, ok := commandHandlers[sa[0]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Unsupported command " + sa[0]}, nil
	}
	return handler(db, sa)
}

func basic(f func(int64, []string) resp.RESP) func(int64, []string) (resp.RESP, connHandler) {
	return func(db int64, sa []string) (resp.RESP, connHandler) {
		return f(db, sa), nil
	}
}
