package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

var echoCommand = "ECHO"

func handleEcho(sa []string, _ int64) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	return &resp.RESPBulkString{Value: sa[1]}, nil
}
