package command

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleGet(sa []string) resp.RESP {
	now := time.Now()
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}
	}
	stateRWMutex.RLock()
	v, ok := state[sa[1]]
	stateRWMutex.RUnlock()
	if !ok {
		return &resp.RESPNull{CompatibilityFlag: 1}
	}
	if v.expiresAt != -1 && v.expiresAt < now.UnixMilli() {
		return &resp.RESPNull{CompatibilityFlag: 1}
	}
	return &resp.RESPBulkString{Value: v.value}
}
