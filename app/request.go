package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleRequest(db int64, r resp.RESP) resp.RESP {
	ra, ok := r.(*resp.RESPArray)
	if !ok {
		return &resp.RESPSimpleError{Value: "Invalid input: expected array"}
	}
	a := make([]string, len(ra.Value))
	for i, v := range ra.Value {
		s, ok := v.(*resp.RESPBulkString)
		if !ok {
			return &resp.RESPSimpleError{Value: "Invalid input: expected array of bulk strings"}
		}
		a[i] = s.Value
	}
	return command.Handle(db, a)
}
