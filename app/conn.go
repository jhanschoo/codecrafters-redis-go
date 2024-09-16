package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/respreader"
)

func handleConn(c net.Conn) error {
	r := respreader.NewBufReader(bufio.NewReader(c))

	for {
		req, err := r.ReadRESP()
		if err != nil {
			fmt.Println("handleConn: error reading input", err)
			return c.Close()
		}
		fmt.Println("handleConn: received request", req.SerializeRESP())
		res := handleRequest(req)
		fmt.Println("handleConn: writing response", res.SerializeRESP())
		c.Write([]byte(res.SerializeRESP()))
	}
}
