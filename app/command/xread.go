package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var xreadCommand = "XREAD"

func handleXread(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) < 4 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 4-element array"}, nil
	}
	subcommand, key, start := sa[1], sa[2], sa[3]
	if strings.ToUpper(subcommand) != "STREAMS" {
		return &resp.RESPSimpleError{Value: "Invalid input: expected STREAMS subcommand"}, nil
	}
	res, err := state.Xread(key, start)
	if err != nil {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return res, nil
}
