package main

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var commandHandlers = map[string]func(sa []string) resp.RESP{
	"PING": handlePing,
	"ECHO": handleEcho,
	"SET":  handleSet,
	"GET":  handleGet,
}

var m = sync.RWMutex{}

var st = make(map[string]string)

func handleCommand(sa []string) resp.RESP {
	if len(sa) == 0 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected non-empty array of bulk strings"}
	}
	handler, ok := commandHandlers[sa[0]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Invalid command"}
	}
	return handler(sa)
}

var pong = &resp.RESPSimpleString{Value: "PONG"}

func handlePing(sa []string) resp.RESP {
	return pong
}

func handleEcho(sa []string) resp.RESP {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}
	}
	return &resp.RESPBulkString{Value: sa[1]}
}

func handleSet(sa []string) resp.RESP {
	if len(sa) != 3 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 3-element array"}
	}
	m.Lock()
	st[sa[1]] = sa[2]
	m.Unlock()
	return &resp.RESPSimpleString{Value: "OK"}
}

func handleGet(sa []string) resp.RESP {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}
	}
	m.RLock()
	v, ok := st[sa[1]]
	m.RUnlock()
	if !ok {
		return &resp.RESPNull{}
	}
	return &resp.RESPBulkString{Value: v}
}
