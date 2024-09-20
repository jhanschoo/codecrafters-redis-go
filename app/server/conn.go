package server

import (
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/command"
	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func HandleConn(c net.Conn) error {
	r := respreader.NewBufferedRESPConnReader(c)
	return HandleReader(r)
}

func HandleReader(r *respreader.BufferedRESPConnReader) error {
	for {
		if err := command.HandleNext(r); err != nil {
			if err == io.EOF {
				log.Println("handleReader: connection closed by client")
			} else {
				log.Println("handleReader: error reading input", err)
			}
			return r.Conn.Close()
		}
	}
}
