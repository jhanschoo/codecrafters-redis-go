package main

import (
	"io"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func handleConn(c net.Conn) error {
	var db int64 = 0
	r := respreader.NewBufReader(c)

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
		res, next := handleRequest(db, req)
		log.Println("handleConn: writing response", strconv.Quote(res.SerializeRESP()))
		c.Write([]byte(res.SerializeRESP()))
		if next != nil {
			(*next)(c)
		}
	}
}
