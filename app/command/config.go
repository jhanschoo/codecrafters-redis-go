package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

var configCommand = "CONFIG"

var configCommandHandlers = map[string]func(sa []string, db int64) (resp.RESP, error){
	"GET": handleConfigGet,
}

func handleConfigCommands(sa []string, db int64) (resp.RESP, error) {
	if len(sa) <= 1 {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG command"}, nil
	}
	handler, ok := configCommandHandlers[sa[1]]
	if !ok {
		return &resp.RESPSimpleError{Value: "Unsupported CONFIG " + sa[1] + " command"}, nil
	}
	return handler(sa, db)

}
