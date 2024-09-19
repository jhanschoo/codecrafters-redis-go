package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

var pingCommand = "PING"

var pong = &resp.RESPSimpleString{Value: "PONG"}

func handlePing(sa []string, ctx Context) (resp.RESP, error) {
	if ctx.IsReplica && ctx.IsPrivileged {
		return nil, nil
	}
	return pong, nil
}
