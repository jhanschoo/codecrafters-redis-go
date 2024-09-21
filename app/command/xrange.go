package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var xrangeCommand = "XRANGE"

func handleXrange(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) != 4 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 4-element array"}, nil
	}
	key, start, end := sa[1], sa[2], sa[3]
	res, err := state.Xrange(key, start, end)
	if err != nil {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return res, nil
}
