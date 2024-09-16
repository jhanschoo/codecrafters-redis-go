package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleSet(sa []string) resp.RESP {
	var key string
	var value string
	var expiresAt int64
	switch len(sa) {
	case 3:
		key = sa[1]
		value = sa[2]
		expiresAt = -1
	case 5:
		if strings.ToUpper(sa[3]) != "PX" {
			return &resp.RESPSimpleError{Value: "Invalid input: expected PX as 4th element"}
		}
		key = sa[1]
		value = sa[2]
		now := time.Now().UnixMilli()
		px, err := strconv.ParseInt(sa[4], 10, 64)
		if err != nil {
			return &resp.RESPSimpleError{Value: "Invalid input: expected integer as 5th element"}
		}
		expiresAt = now + px
	default:
		return &resp.RESPSimpleError{Value: "Invalid input: expected 3 or 5-element array"}
	}
	v := kvValue{value: value, expiresAt: expiresAt}
	stateRWMutex.Lock()
	state[key] = v
	stateRWMutex.Unlock()
	return &resp.RESPSimpleString{Value: "OK"}
}
