package server

import (
	"bufio"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func HandleConn(c net.Conn, br *bufio.Reader, isPrivileged bool) error {
	var db int64 = 0
	isReplica := replication.GetReplicationInfo().Role == "slave"
	if isReplica && isPrivileged {
		log.Println("listening to master as a replica")
	}
	if br == nil {
		br = bufio.NewReader(c)
	}
	r := respreader.NewBufReader(br)
	ch := command.GetDefaultHandler()

	for {
		req, err := r.ReadRESP()
		if err != nil {
			if err == io.EOF {
				log.Println("handleConn: connection closed by client")
			} else {
				log.Println("handleConn: error reading input", err)
			}
			return c.Close()
		}
		log.Println("handleConn: received request", strconv.Quote(req.SerializeRESP()))
		ctx := command.Context{
			Conn:                    c,
			Db:                      db,
			IsReplica:               isReplica,
			IsPrivileged:            isPrivileged,
			ExecuteAndWriteToSlaves: executeAndWriteToSlavesReplica,
		}
		if !isReplica {
			ctx.ExecuteAndWriteToSlaves = executeAndWriteToSlaves
		}
		err = ch.Do(req, ctx)
		if err != nil {
			log.Println("handleConn: error handling request", err)
			return c.Close()
		}
	}
}

func executeAndWriteToSlavesReplica(f func() error, _ []string) {
	f()
}

func executeAndWriteToSlaves(f func() error, sa []string) {
	ba := []byte(resp.EncodeStringSlice(sa).SerializeRESP())
	replication.ExecuteAndWriteToListenersAtomically(f, ba)
}
