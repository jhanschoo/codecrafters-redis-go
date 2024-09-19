package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

var pingCommand = "PING"

var pong = &resp.RESPSimpleString{Value: "PONG"}

func handlePing(sa []string, _ int64) (resp.RESP, error) {
	return pong, nil
}
