package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleConfigGet(sa []string) resp.RESP {
	ra := make([]resp.RESP, 0, (len(sa)-2)*2)
	// This handles only the special case where len(sa) == 3 and the sole argument matckes a config key.
	v, ok := config.Get(sa[2])
	if !ok {
		return &resp.RESPNull{CompatibilityFlag: 1}
	}
	ra = append(ra, &resp.RESPBulkString{Value: sa[2]}, &resp.RESPBulkString{Value: v})
	// RESP2 compatible response
	return &resp.RESPArray{Value: ra}
	// RESP3 compatible response
	// return &resp.RESPMap{Value: ra}
}
