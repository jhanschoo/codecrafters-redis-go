package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleConfigGet(sa []string, ctx Context) (resp.RESP, error) {
	// This handles only the special case where len(sa) == 3 and the sole argument matckes a config key.
	keys := config.Keys()
	for _, k := range keys {
		if k == sa[2] {
			return resp.EncodeStringSlice([]string{sa[2], config.Get(sa[2])}), nil
		}
	}
	return respNull, nil
}
