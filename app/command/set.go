package command

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

func handleSet(db int64, sa []string) resp.RESP {
	var (
		key   string
		value string
		px    int64
		err   error
	)
	switch len(sa) {
	case 3:
		key = sa[1]
		value = sa[2]
		px = -1
	case 5:
		if strings.ToUpper(sa[3]) != "PX" {
			return &resp.RESPSimpleError{Value: "Invalid input: expected PX as 4th element"}
		}
		key = sa[1]
		value = sa[2]
		px, err = strconv.ParseInt(sa[4], 10, 64)
		if err != nil {
			return &resp.RESPSimpleError{Value: "Invalid input: expected integer as 5th element"}
		}
	default:
		return &resp.RESPSimpleError{Value: "Invalid input: expected 3 or 5-element array"}
	}
	state.Set(db, key, value, px)
	return &resp.RESPSimpleString{Value: "OK"}
}
