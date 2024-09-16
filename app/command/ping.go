package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

var pong = &resp.RESPSimpleString{Value: "PONG"}

func handlePing(sa []string) resp.RESP {
	return pong
}
