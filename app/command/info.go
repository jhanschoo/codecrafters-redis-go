package command

import "github.com/codecrafters-io/redis-starter-go/app/resp"

func handleInfo(_ int64, sa []string) resp.RESP {
	if len(sa) != 2 || sa[1] != "replication" {
		return &resp.RESPSimpleError{Value: `Unsupported input: only INFO replication is supported for the INFO command`}
	}
	return &resp.RESPBulkString{Value: "role:master\r\n"}
}
