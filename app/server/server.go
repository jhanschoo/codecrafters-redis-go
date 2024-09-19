package server

import (
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

func Serve() {
	port, _ := config.Get("port")

	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Println("Failed to bind to port", port)
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go HandleConn(c)
	}
}
