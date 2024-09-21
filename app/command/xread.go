package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var xreadCommand = "XREAD"

func handleXread(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) < 4 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected at least 4-element array"}, nil
	}
	var blockTimeout int64 = -1
	var err error
	i := 1
	for ; i < len(sa); i++ {
		shouldBreak := false
		switch strings.ToUpper(sa[i]) {
		case "STREAMS":
			i++
			shouldBreak = true
		case "BLOCK":
			i++
			if i >= len(sa) {
				return &resp.RESPSimpleError{Value: "Invalid input: expected a block timeout value after BLOCK"}, nil
			}
			blockTimeout, err = strconv.ParseInt(sa[i], 10, 64)
			if err != nil {
				return &resp.RESPSimpleError{Value: "Invalid input: expected an integer as block timeout value after BLOCK"}, nil
			}
		default:
			return &resp.RESPSimpleError{Value: "Invalid input: expected STREAMS or BLOCK"}, nil
		}
		if shouldBreak {
			break
		}
	}

	kids := sa[i:]
	res, err := state.Xread(kids, time.Duration(blockTimeout)*time.Millisecond)
	if err != nil {
		return &resp.RESPSimpleError{Value: err.Error()}, nil
	}
	return res, nil
}
