package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func handleConn(c net.Conn) error {
	buf := make([]byte, 1024)
	p := parser.NewParser()
	for n, err := c.Read(buf); err != io.EOF; n, err = c.Read(buf) {
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			os.Exit(1)
		}
		subbuf := buf[:n]
		for len(subbuf) > 0 {
			fmt.Printf("Received %d bytes: %s\n", len(subbuf), subbuf)
			r, j, err := p.Parse(subbuf, 0)
			// handle error
			if err != nil {
				fmt.Println("Error parsing input: ", err.Error())
				return c.Close()
			}
			subbuf = subbuf[j:]

			// discrepancy found in testing: we should clobber the buffer if we have a result
			subbuf = subbuf[:0]
			// handle complete parsing
			if r != nil {
				r_res := handleRequest(r)
				c.Write(r_res.Serialize())
			}
		}
	}
	return c.Close()
}

func handleRequest(r resp.RESP) resp.RESP {
	a, ok := r.(*resp.RESPArray)
	if !ok {
		return &resp.RESPError{Value: "Invalid input: expected array"}
	}
	if len(a.Value) == 0 {
		return &resp.RESPError{Value: "Invalid input: expected non-empty array"}
	}
	cmd, ok := a.Value[0].(*resp.RESPBulkString)
	if !ok {
		return &resp.RESPError{Value: "Invalid input: expected command to be a bulk string"}
	}
	switch string(cmd.Value) {
	case "PING":
		return handlePing(a)
	case "ECHO":
		return handleEcho(a)
	default:
		return &resp.RESPError{Value: "Invalid input: unknown command"}
	}
}

func handlePing(_ *resp.RESPArray) resp.RESP {
	return &resp.RESPSimpleString{Value: "PONG"}
}

func handleEcho(a *resp.RESPArray) resp.RESP {
	if len(a.Value) != 2 {
		return &resp.RESPError{Value: "Invalid input: expected 2-element array"}
	}
	s, ok := a.Value[1].(*resp.RESPBulkString)
	if !ok {
		return &resp.RESPError{Value: "Invalid input: expected second element to be a bulk string"}
	}
	return s
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(c)
	}

}
