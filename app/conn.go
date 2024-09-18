package main

import (
	"bufio"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func handleConn(c net.Conn) error {
	r := respreader.NewBufReader(bufio.NewReader(c))

	for {
		req, err := r.ReadRESP()
		if err != nil {
			log.Println("handleConn: error reading input", err)
			return c.Close()
		}
		log.Println("handleConn: received request", strconv.Quote(req.SerializeRESP()))
		res := handleRequest(req)
		log.Println("handleConn: writing response", strconv.Quote(res.SerializeRESP()))
		c.Write([]byte(res.SerializeRESP()))
	}
}
