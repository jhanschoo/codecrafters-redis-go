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

func handleConn(c net.Conn) error {
	var db int64 = 0
	r := respreader.NewBufReader(bufio.NewReader(c))
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
		log.Println("handleConn: received request", strconv.Quote(req.SerializeRESP()), "delegating to handleRequest")
		err = ch.Do(req, command.Context{Conn: c, Db: db, WriteToSlaves: writeToSlaves})
		if err != nil {
			log.Println("handleConn: error handling request", err)
			return c.Close()
		}
	}
}

func writeToSlaves(sa []string) error {
	ba := []byte(resp.EncodeStringSlice(sa).SerializeRESP())
	replication.WriteToAllListeners(ba)
	return nil
}
