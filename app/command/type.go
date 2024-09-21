package command

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/state"
)

var typeCommand = "TYPE"

var (
	stringTypeResp = &resp.RESPSimpleString{Value: "string"}
	listTypeResp   = &resp.RESPSimpleString{Value: "list"}
	setTypeResp    = &resp.RESPSimpleString{Value: "set"}
	zsetTypeResp   = &resp.RESPSimpleString{Value: "zset"}
	hashTypeResp   = &resp.RESPSimpleString{Value: "hash"}
	streamTypeResp = &resp.RESPSimpleString{Value: "stream"}
	noneTypeResp   = &resp.RESPSimpleString{Value: "none"}
)

func handleType(sa []string, _ Context) (resp.RESP, error) {
	if len(sa) != 2 {
		return &resp.RESPSimpleError{Value: "Invalid input: expected 2-element array"}, nil
	}
	key := sa[1]
	_, ok := state.Get(key)
	if !ok {
		return noneTypeResp, nil
	}
	return stringTypeResp, nil
}
