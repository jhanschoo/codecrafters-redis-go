package server

import (
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func HandleConn(c net.Conn) error {
	r := respreader.NewBufferedRESPConnReader(c)
	return HandleConnReader(r)
}

func HandleConnReader(r *respreader.BufferedRESPConnReader) error {
	// Warning: opts here is reused
	queued := resp.NewComSlice()
	for {
		if err := command.HandleNext(r, command.HandlerOptions{
			Queued:        queued,
			InTransaction: false,
		}); err != nil {
			if err == io.EOF {
				log.Println("handleReader: connection closed by client")
			} else {
				log.Println("handleReader: error reading input", err)
			}
			return r.Conn.Close()
		}
	}
}
