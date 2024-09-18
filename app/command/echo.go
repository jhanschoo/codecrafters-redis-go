package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

func handleEcho(_ int64, sa []string) resp.RESP {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}
	}
	return &resp.RESPBulkString{Value: sa[1]}
}
