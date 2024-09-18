package command

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleInfo(_ int64, sa []string) resp.RESP {
	if len(sa) != 2 || sa[1] != "replication" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only INFO replication is supported for the INFO command`}
	}
	role := "master"
	if replicaof, _ := config.Get("replicaof"); replicaof != "" {
		role = "slave"
	}
	return &resp.RESPBulkString{Value: fmt.Sprintf("role:%v\r\n", role)}
}
