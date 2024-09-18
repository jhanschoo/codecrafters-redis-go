package main

import (
	"log"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/daemon"
)

func main() {
	log.Println("Logs from your program will appear here!")

	config.ParseConfig()

	daemon.InitializeDaemons()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(c)
	}

}
