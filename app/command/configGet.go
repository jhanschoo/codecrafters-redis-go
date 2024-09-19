package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleConfigGet(sa []string, _ int64) (resp.RESP, error) {
	// This handles only the special case where len(sa) == 3 and the sole argument matckes a config key.
	v, ok := config.Get(sa[2])
	if !ok {
		return &resp.RESPNull{CompatibilityFlag: 1}, nil
	}
	return resp.EncodeStringSlice([]string{sa[2], v}), nil
}
