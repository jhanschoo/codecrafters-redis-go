package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func handleConn(c net.Conn) error {
	buf := make([]byte, 1024)
	for _, err := c.Read(buf); err != io.EOF; _, err = c.Read(buf) {
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			os.Exit(1)
		}
		c.Write([]byte("+PONG\r\n"))
	}
	return c.Close()
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
